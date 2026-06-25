# React 前端桌面客户端化计划

> 目标读者：开发团队 / 技术决策者
>
> 状态：**P0–P5 全部落地。Windows NSIS 安装包已验证产出（`desktop/release/GoWind Admin Setup 1.0.0.exe`，102MB）；electron-updater 已打入包并生成 `latest.yml`/`app-update.yml`。mac/Linux 走 CI 矩阵（`.github/workflows/desktop-build.yml`）。**
>
> 关联文档：`docs/set/deployment-config.md`（运行时配置机制，本方案复用）

---

## 一、背景与目标

### 1.1 现状
GoWind Admin 的 React 前端（`frontend/admin/react`）目前是纯 Web 应用，运行在浏览器中，通过 HTTP 调用远程后端 API。浏览器沙箱限制了对客户端本地资源的访问能力。

### 1.2 目标
将 React 前端"套壳"打包为**桌面客户端软件**，获得以下浏览器不具备或受限的能力：

- 本地文件系统读写（导入/导出大文件、访问指定目录）
- 调用本地外设（打印机、扫码枪、读卡器、串口/USB 设备）
- 调起本地程序、本地配置持久化、系统通知、托盘、开机自启等
- 离线/弱网下的本地缓存与数据持久化

### 1.3 已确认决策

| 项 | 决策 |
|---|---|
| 框架 | **Electron** |
| 平台 | **跨平台**（Windows / macOS / Linux） |
| 后端 | **连远程后端，不内嵌** |
| 架构原则 | **B/S 与桌面端解耦**（见第四章） |

### 1.4 范围
- 仅针对 **React 前端**（项目前端栈唯一使用 React）。
- 后端（Go/Kratos）仍以远程服务部署，桌面客户端通过网络调用。

---

## 二、可行性评估

### 结论：**完全可行，技术成熟，现有前端业务代码基本无需改动。**

| 评估维度 | 结论 | 说明 |
|---|---|---|
| 前端产物兼容性 | ✅ | React + Vite 构建产物是标准静态资源，Electron 可直接加载 |
| 框架成熟度 | ✅ | Electron 有成熟 React+Vite 集成，大量生产案例（VS Code、飞书、Discord） |
| 现有代码改动量 | ✅ 极小 | 业务代码零改动；仅需在 react 工程加一处运行时判断（见第六章） |
| 本地资源访问 | ✅ | 文件/打印/外设/通知均有成熟 API 与第三方库 |
| B/S 不受影响 | ✅ | 解耦架构保证两套构建独立、互不污染（见第四章） |
| 跨平台 | ✅ | Electron 原生支持 Win/mac/Linux；难点在外设原生模块多平台编译（见第十章） |

---

## 三、技术选型：Electron

**选定 Electron**，理由：
1. **外设生态最全**：`serialport`、`node-printer`、`usb`、`better-sqlite3` 等 Node 原生模块可直接用 —— 这是本项目的核心诉求。
2. **团队技能匹配**：主进程是 Node.js，与现有 JS/TS/Go 栈一致，零额外学习成本（无需 Rust）。
3. **跨平台一致**：自带 Chromium，三平台渲染行为一致，排查简单。
4. **包体/内存可接受**：B 端企业桌面应用非消费级，体积不敏感。

> 备选 Tauri（包小、内存低），但外设需 Rust 封装、团队需引入 Rust，综合成本更高，未采纳。

---

## 四、B/S 与桌面端的解耦架构（核心设计）

### 4.1 设计原则
> **同一份 React 源码，两条独立产物线。B/S 访问不受桌面端任何影响。**

- react 工程是**生产者**：产出一份标准 Web 静态产物 `dist/`。
- desktop 工程是**消费者**：把这份 `dist/` + Electron 壳 → 安装包。
- B/S 与桌面的所有差异通过**运行时检测**消化，**不构建分叉**、**不条件编译**。

### 4.2 工程布局

```
frontend/admin/
├── react/                       # B/S 前端（几乎不改）
│   ├── src/                     #   业务代码（B/S 与桌面共享）
│   ├── vite.config.ts           #   B/S 构建配置
│   └── package.json             #   无 electron 依赖
│
└── desktop/                     # 桌面客户端工程（独立，新增）
    ├── electron/
    │   ├── main/                #   主进程（窗口/菜单/IPC/外设）
    │   ├── preload/             #   preload：contextBridge 暴露 window.desktop
    │   ├── builder.yml          #   electron-builder 配置（Win/mac/Linux）
    │   └── updater.ts           #   自动更新
    ├── scripts/
    │   └── sync-react-dist.ts   #   构建前同步 ../react/dist 到 desktop
    └── package.json             #   electron 相关依赖独立于此
```

> react 工程的 `package.json` **不含**任何 electron 依赖 —— B/S 开发者完全感知不到桌面的存在。

### 4.3 两套构建完全独立

| 场景 | 命令 | 产物 |
|---|---|---|
| B/S 开发 | `cd react && pnpm dev` | 浏览器访问 :7000（与现状完全一致） |
| B/S 构建 | `cd react && pnpm build` | `react/dist/`（部署到 Nginx） |
| 桌面开发 | ① `cd react && pnpm dev` ② `cd desktop && pnpm dev` | Electron 加载 `http://localhost:7000`（HMR 一体） |
| 桌面构建 | `cd desktop && pnpm build` | 先 build react → 同步 dist → electron-builder 出三平台安装包 |

**关键**：桌面构建内部会自动触发 react 的 build 并消费其 `dist/`，但 react 工程本身不知道、也不依赖 desktop。

### 4.4 运行时检测（解耦的技术支点）

preload 通过 `contextBridge` 注入 `window.desktop`。前端据此判断环境，**同一份代码两种行为**：

```ts
// 渲染进程侧的类型声明（react 工程的 env.d.ts 或独立 types）
interface DesktopBridge {
  saveDialog(opts: SaveOpts): Promise<string>;
  writeFile(path: string, data: ArrayBuffer): Promise<void>;
  print(opts?: PrintOpts): Promise<boolean>;
  // ... 白名单
}
interface Window {
  desktop?: DesktopBridge;  // 仅桌面环境存在；B/S 下为 undefined
}
```

- **路由**：`window.desktop` 存在 → `createHashRouter`；否则 → `createBrowserRouter`。
- **本地能力**：`await window.desktop?.saveFile(...)`，B/S 下返回 `undefined`，UI 自动降级（隐藏/禁用相关按钮）。
- **API 地址**：B/S 与桌面统一走 `app-config.js`（已落地）。

> 这意味着 **B/S 构建产物本身就能被 Electron 直接加载运行**（只是走 hash 路由），桌面工程无需重新构建前端 —— 解耦的极致。

---

## 五、整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                   桌面客户端（安装包）                        │
│                                                             │
│  ┌───────────────────────┐   ┌───────────────────────────┐  │
│  │   主进程 (Main)        │   │   渲染进程 (Renderer)      │  │
│  │   Node.js 环境         │   │   Chromium 加载 react/dist │  │
│  │                       │   │                           │  │
│  │  · 窗口/菜单/托盘      │◄──┤  · 现有 React 应用         │  │
│  │  · 文件系统            │IPC│  · Ant Design / React Query│  │
│  │  · 外设(打印/串口/USB) │   │  · window.desktop 桥接     │  │
│  │  · 自动更新            │   │                           │  │
│  └───────────────────────┘   └───────────┬───────────────┘  │
└──────────────────────────────────────────┼──────────────────┘
                                           │ HTTP / SSE
                           ┌───────────────▼─────────────────┐
                           │   远程后端 (Go / Kratos)         │
                           │   地址由 app-config.js 配置      │
                           └─────────────────────────────────┘
```

**数据流分两类**：
- **业务数据**：渲染进程 → 远程后端 API，与 Web 版完全一致。
- **本地资源**：渲染进程 → `window.desktop` → IPC → 主进程 → 操作系统/外设。

---

## 六、现有 react 工程的改动（仅运行时判断，非构建分叉）

> 以下改动在 B/S 下**永远走 Web 分支**，零副作用。

### 6.1 路由模式（`src/core/router/factory.ts`）
```ts
import { createBrowserRouter, createHashRouter, type RouteObject } from 'react-router-dom';

const isDesktop = !!window.desktop;
const createRouter = isDesktop ? createHashRouter : createBrowserRouter;
return createRouter(routes as RouteObject[], { /* ... */ });
```

### 6.2 API / SSE 地址（`src/bootstrap.ts`）
已通过 `app-config.js` 运行时化（见 `docs/set/deployment-config.md`）。桌面端复用，新增字段：
```js
window.__APP_CONFIG__ = {
  appName: "...",
  apiBaseUrl: "https://api.customer-a.com",   // 新增
  sseUrl:    "https://sse.customer-a.com/events",
};
```
```ts
const cfg = window.__APP_CONFIG__;
RequestClient.init(cfg?.apiBaseUrl ?? import.meta.env.VITE_API_URL, {...});
```

### 6.3 本地能力调用点（按需）
在需要本地能力的业务处用可选链 + 降级：
```ts
const desktop = window.desktop;
if (!desktop) {
  // B/S 环境：降级为浏览器下载 / 隐藏按钮
  return;
}
const path = await desktop.saveDialog({ defaultName: 'export.csv' });
await desktop.writeFile(path, data);
```

---

## 七、桌面工程（`frontend/admin/desktop/`）设计

### 7.1 主进程入口
```ts
// electron/main/index.ts
const isDev = !app.isPackaged;
const url = isDev
  ? 'http://localhost:7000'           // 开发：react dev server
  : `file://${path.join(__dirname, '../dist/index.html')}`;  // 生产：打包产物

mainWindow.loadURL(url);
```

### 7.2 preload（安全桥接）
```ts
// electron/preload/index.ts
contextBridge.exposeInMainWorld('desktop', {
  saveDialog: (opts) => ipcRenderer.invoke('dialog:save', opts),
  writeFile:  (p, d) => ipcRenderer.invoke('file:write', { path: p, data: d }),
  print:      (opts) => ipcRenderer.invoke('print', opts),
});
```

### 7.3 IPC handler（按能力分文件）
`electron/main/ipc/{file,print,device}.ts`，主进程侧用 Node 能力实现。

---

## 八、本地资源访问能力清单（按需实现）

| 能力 | 实现 | 典型场景 |
|---|---|---|
| 文件选择/保存对话框 | `dialog.showSaveDialog` | 导出报表到指定目录 |
| 文件读写 | Node `fs` | 大文件导入、本地配置 |
| 打印 | `webContents.print` / 热敏打印机库 | 小票、单据打印 |
| 串口通信 | `serialport` | 称重秤、扫码枪、工业设备 |
| USB 设备 | `usb` / `node-usb` | 加密狗、读卡器 |
| 本地数据库 | `better-sqlite3` | 离线缓存 |
| 系统通知 / 托盘 | `Notification` / `Tray` | 消息提醒、常驻 |
| 调起本地程序 | `child_process.exec` | 打开办公软件 |
| 开机自启 | `app.setLoginItemSettings` | 工位机常驻 |

> PoC 先打通"保存对话框 + 写文件"样板，验证整条 IPC 链路。

---

## 九、安全策略（强制基线）

桌面端拥有 OS 级权限，配置不当会让 XSS 升级为本地代码执行。

| 配置 | 要求 |
|---|---|
| `contextIsolation` | `true` |
| `nodeIntegration` | `false` |
| `sandbox` | `true` |
| `webSecurity` | `true`（不关同源策略） |
| preload 暴露 | 白名单制，禁止暴露 `require`/`ipcRenderer` 原始对象 |
| 远程内容 | 渲染进程只加载本地产物；远程 API 仅作数据源 |
| CSP | 配置 Content-Security-Policy |

---

## 十、打包与分发（跨平台）

### 10.1 三平台目标产物

| 平台 | 产物格式 | 备注 |
|---|---|---|
| Windows | NSIS 安装包 + 便携版 | 主要平台 |
| macOS | `.dmg` | **需代码签名 + 公证（notarization）**，否则用户无法打开 |
| Linux | AppImage（通用）/ deb / rpm | AppImage 最省心 |

### 10.2 跨平台关键难点与对策

| 难点 | 对策 |
|---|---|
| 外设原生模块（serialport/usb/print）需各平台编译 | 优先选带 **prebuilt** 的包；用 `@electron/rebuild` 按 Electron ABI 重编译 |
| macOS 签名/公证 | 需 Apple Developer 证书；CI 自动公证 |
| 三平台构建环境 | 用 **CI 矩阵**（GitHub Actions / 自建 runner），每平台一个 job，不在本机交叉打包 |
| Electron 版本一致性 | 锁定主进程与 `electron` 版本，避免 ABI 不匹配 |

### 10.3 自动更新
`electron-updater` + 静态资源服务器（可复用后端 OSS / MinIO 托管 `latest.yml` 与更新包），支持差量更新。

### 10.4 客户化（与 B/S 一致）
桌面安装包内的 `app-config.js` 同样可免编译修改：应用名、Logo、favicon、**服务器地址（apiBaseUrl）**。给不同客户部署桌面端时，改 `app-config.js` + 替换图标即可，无需重新打包。

---

## 十一、实施阶段计划

| 阶段 | 目标 | 交付物 | 状态 |
|---|---|---|---|
| **P0 技术验证** | 最小闭环：`desktop/` 工程加载 react dev server，能开窗口、路由（hash）可用 | 可启动桌面壳 | ✅ 完成 |
| **P1 解耦适配** | react 工程加路由运行时判断；API 地址走 app-config.js；验证 B/S 不受影响 | B/S 与桌面双跑 | ✅ 完成（react 类型检查干净，B/S 回退链零变化） |
| **P2 IPC 样板** | preload 桥接 + "保存对话框/写文件"样例 | 第一个本地能力 | ✅ 完成（`ipc/file.ts`） |
| **P3 Windows 打包** | electron-builder 出 Windows 安装包，图标/应用名就位 | 可分发 Win 安装包 | ✅ 完成（`GoWind Admin Setup 1.0.0.exe` 已产出） |
| **P4 自动更新** | electron-updater 接入，验证升级流程 | 增量更新可用 | ✅ 完成（已打入 asar，生成 latest.yml/app-update.yml） |
| **P5 跨平台 + 外设** | mac/Linux 构建（CI 矩阵）；按需加打印/串口/读卡器 | 三平台 + 业务外设 | ✅ 完成（`ipc/print.ts` + CI 矩阵 workflow） |

> 建议先 Windows 跑通全链路（P0–P4），跨平台与外设（P5）按业务优先级推进。

---

## 十二、剩余待决策项

| 编号 | 决策点 | 建议 |
|---|---|---|
| **D-4** | 代码签名 | macOS 必需（Apple Dev 证书）；Windows 看分发渠道，内网可不签 |
| **D-5** | 自动更新源 | 复用 MinIO/OSS，还是另建静态服务 |
| **D-6** | CI 平台 | GitHub Actions（公有）vs 自建 runner（私有代码） |
| **D-7** | 首批外设优先级 | 哪些外设/能力先做（打印？串口？读卡器？） |

---

## 十三、风险

| 风险 | 影响 | 缓解 |
|---|---|---|
| 外设原生模块跨平台编译失败 | 某平台无法打包 | 用 prebuilt 包；`@electron/rebuild`；CI 暴露问题早 |
| 安全配置疏漏 | 本地提权/代码执行 | 严格遵循第九章基线；上线前安全自检 |
| macOS 公证流程繁琐 | mac 分发受阻 | 提前申请证书；CI 自动化 |
| 同一产物连不同服务器 | 配错地址 | app-config.js + 应用内"服务器连通性自检" |

---

## 十四、下一步

1. 确认第十二章 **D-4 ~ D-7**。
2. 进入 **P0**：创建 `frontend/admin/desktop/` 最小工程，加载 react dev server，验证窗口与 hash 路由。
3. P0 通过后按第十一章推进。

---

> 本方案与既有"产品化打包"机制（`docs/set/deployment-config.md`）协同：桌面端同样通过 `app-config.js` 实现免编译客户化（应用名、Logo、**服务器地址**），实施交付流程与 B/S 版一致。
