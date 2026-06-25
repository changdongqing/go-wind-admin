# GoWind Admin 桌面客户端（Electron 壳）

> 这是 react 前端的**桌面端消费者**，不含业务逻辑。业务代码全在 `../react/`。
>
> 架构原则：**B/S 与桌面解耦**。详见 `../../../docs/desktop/desktop-client-plan.md`。

## 工程关系

- `../react/` 生产 `dist/`（标准 Web 产物，零桌面耦合）
- 本工程消费该产物：开发态加载 react dev server，生产态把 react dist 打进安装包
- react 工程感知桌面仅靠运行时判断 `window.desktop`（B/S 下永远不存在，零副作用）

## 开发（需两个终端）

**终端 1** — 启动 react dev server
```bash
cd ../react
pnpm dev          # 默认 http://localhost:7000
```

**终端 2** — 启动 electron（加载上面的 dev server）
```bash
cd desktop
pnpm dev
```

> react 代码改动由 vite HMR 热更；electron 侧 TS 改动需重启 `pnpm dev`。

## 链路自检

应用启动后，在 electron 窗口的 DevTools Console 执行：

```js
// P0：IPC 链路
await window.desktop.ping()                                      // 'pong'
window.desktop.isDesktop                                          // true，前端据此切换 hash 路由

// P2：本地文件能力（保存对话框 + 写文件）
const p = await window.desktop.saveDialog({ defaultName: 'hello.txt' })  // 弹框选路径，取消返回 ''
if (p) await window.desktop.writeFile(p, new TextEncoder().encode('你好，桌面端'))  // 写入文件

// P5：通用打印（打印当前窗口页面）
await window.desktop.print({ printBackground: true })             // 弹打印机对话框；silent: true 直接用默认打印机

// P4：自动更新（通常由 app-config.js 配 updateServerUrl，启动时自动检查；此处手动触发）
// await window.desktop.checkForUpdates('https://updates.example.com/desktop/')
```

> B/S（浏览器）下 `window.desktop` 为 `undefined`，前端用 `window.desktop?.xxx` 可选链调用即可自动降级。

## 自动更新（P4）

- 依赖 `electron-updater`（已声明于 `dependencies`）。更新地址运行时取自 `app-config.js` 的 `updateServerUrl`（解耦：免编译改地址）。
- 启动时若检测到 `updateServerUrl`，自动检查 → 后台下载 → 下载完成弹窗询问是否重启安装。
- 开发期自检：本地起静态服务托管「伪造更新包 + `latest.yml`」，配置 `dev-app-update.yml` 的 `url` 指向它。
- ⚠️ Windows 生产环境的完整自动更新依赖代码签名（D-4）；未签名时仅开发期可验证。

## 跨平台 CI（P5）

三平台构建发布走 CI 矩阵（不在本机交叉打包），模板见仓库根 `.github/workflows/desktop-build.yml`（tag `v*` 触发 / 手动触发）。凭据全部走仓库 Secrets，详见 `docs/caozuo/desktop-optional-config.md` 第三章。


## 构建（生产安装包）

`pnpm build` 自动跑完整条链路：构建 react（相对 base）→ 同步到 `renderer-dist/` → 编译主进程 → electron-builder 出包，产物在 `release/`。

```bash
cd desktop
pnpm build            # 当前平台
pnpm build:win        # Windows (NSIS 安装包)
pnpm build:mac        # macOS (DMG)   —— 需在 macOS 上执行
pnpm build:linux      # Linux (AppImage)
```

**为什么桌面构建要重新构建 react（`--base=./`）**：Electron 用 `file://` 加载打包产物，Vite 默认 `base: '/'` 的绝对路径会指向磁盘根而 404。桌面构建以相对 base 重出一份产物，再由 `sync:react` 把 `index.html` 里 public 资源（`app-config.js`/`favicon.ico`）的根绝对路径改写为相对路径。react 源码与 B/S 产物**完全不受影响** —— B/S 仍用自己的 `pnpm build`（base '/'）。

**分步执行**（调试用）：
```bash
pnpm build:react      # 构建 react（相对 base）→ ../react/dist
pnpm sync:react       # 同步 + 改写路径 + 生成占位图标 → renderer-dist
pnpm build:electron   # 编译主进程/preload → dist-electron
npx electron-builder --win   # 仅打包（跳过前端重建，前提是已 sync:all 过）
```

**国内/网络受限环境**：electron-builder 首次运行要从 GitHub 下载 Electron 二进制与 NSIS 等打包工具，易超时。设置镜像后再构建（bash）：
```bash
export ELECTRON_MIRROR="https://npmmirror.com/mirrors/electron/"
export ELECTRON_BUILDER_BINARIES_MIRROR="https://npmmirror.com/mirrors/electron-builder-binaries/"
pnpm build:win
```
> 下载缓存在 `%LOCALAPPDATA%\electron-builder\cache`，之后构建可免镜像。

> Windows 安装器图标取自 `build/icon.ico`（首次构建自动从 `../react/public/favicon.ico` 种子生成，可替换为 256×256+ 正式图标）。
> 跨平台构建建议用 CI 矩阵，不在本机交叉打包。详见 `docs/caozuo/desktop-optional-config.md` 第三章。

## 可选配置

代码签名 / 自动更新源 / CI 平台 / 外设启用：见 `docs/caozuo/desktop-optional-config.md`。

## 目录结构

```
desktop/
├── electron/
│   ├── main/              # 主进程（窗口/IPC 注册/外设）
│   │   ├── index.ts       #   入口：窗口 + 注册全部 IPC + 初始化自动更新
│   │   ├── updater.ts     #   自动更新（electron-updater，update:* IPC）
│   │   └── ipc/           #   IPC handler（按能力分文件）
│   │       ├── desktop.ts #     ping 链路自检
│   │       ├── file.ts    #     保存对话框 + 写文件
│   │       ├── print.ts   #     通用打印（webContents.print）
│   │       └── index.ts   #     注册聚合
│   └── preload/           # preload：contextBridge 暴露 window.desktop（白名单）
├── scripts/
│   └── sync-react-dist.mjs # 构建前同步 ../react/dist → renderer-dist + 改写路径 + 种子图标
├── dev-app-update.yml     # 开发期自动更新自检配置（生产态读 app-update.yml）
├── electron-builder.yml   # 打包配置（含 publish 自动更新元数据）
├── tsconfig.json
└── package.json           # electron 相关依赖独立于此工程
```
