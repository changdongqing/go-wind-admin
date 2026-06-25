# 桌面客户端可选配置操作手册

> 适用对象：**开发 / 运维人员**
>
> 本手册覆盖桌面客户端（`frontend/admin/desktop/`）的四项可选配置。每项均设计为**可配置、可切换**——不写死决策，按实际环境启用即可。
>
> 关联：计划文档 `docs/desktop/desktop-client-plan.md`

---

## 一、代码签名（D-4）

### 作用
给安装包数字签名，避免 Windows SmartScreen 警告、macOS Gatekeeper 拦截。**不签名也能正常打包**，只是用户首次安装时会有风险提示（内网分发通常可接受）。

### 可配置点
electron-builder 通过**环境变量**自动识别签名凭据，无需改代码。配了就签名，不配就跳过。

### 1.1 不签名（默认 / 内网分发）
无需任何配置，直接 `pnpm build` 即可。安装时系统会提示"未知发布者"，点"仍要运行"继续。

### 1.2 Windows 代码签名
**准备**：向证书颁发机构（CA）购买代码签名证书（`.pfx` 文件）。

**配置**（构建前设置环境变量）：
```bash
# PowerShell
$env:CSC_LINK="C:\path\to\cert.pfx"
$env:CSC_KEY_PASSWORD="你的证书密码"
pnpm build:win
```
```bash
# bash
CSC_LINK="/path/to/cert.pfx" \
CSC_KEY_PASSWORD="你的证书密码" \
pnpm build:win
```
> `CSC_LINK` 也接受 Base64 编码的证书内容（适合 CI 环境，把证书存为 secret）。

### 1.3 macOS 签名 + 公证（必做，否则用户打不开）
**准备**：Apple Developer 账号，创建"Developer ID Application"证书（导入钥匙串）+ App 专用密码（用于公证）。

**配置**（构建前设置环境变量）：
```bash
export APPLE_ID="你的Apple ID邮箱"
export APPLE_APP_SPECIFIC_PASSWORD="App专用密码"   # appleid.apple.com 生成
export APPLE_TEAM_ID="你的团队ID"                   # Developer 账号里查
pnpm build:mac
```
electron-builder 会自动完成"签名 + 公证（notarize）+ 装订（staple）"全流程。

### 1.4 启用/禁用切换
| 场景 | 做法 |
|---|---|
| 临时禁用签名 | 不设置上述环境变量即可 |
| 永久关闭 mac 公证 | 在 `electron-builder.yml` 设 `mac.notarize: false` |
| 永久关闭 win 签名 | 在 `electron-builder.yml` 设 `win.signAndEditExecutable: false` |

---

## 二、自动更新源（D-5）

### 作用
发布新版本后，已安装的客户端自动检测并升级（`electron-updater`）。

### 可配置点
更新服务器地址通过**环境变量** `UPDATE_SERVER_URL` 注入，构建时写入产物；运行时从 `app-config.js` 读取（可免编译修改）。

### 2.1 三种更新源可选

| 源 | 适用 | 说明 |
|---|---|---|
| **MinIO / OSS 静态托管**（推荐） | 私有 / 企业内 | 复用后端现有对象存储，最省事 |
| 自建 Nginx 静态目录 | 自有服务器 | 起一个静态文件目录服务 |
| GitHub Releases | 开源 / 公有 | 配置最简，但仓库需公开或用 token |

### 2.2 配置更新地址
构建时注入（环境变量）：
```bash
export UPDATE_SERVER_URL="https://updates.customer-a.com/desktop/"
pnpm build
```
或在部署后改 `app-config.js`（免编译）：
```js
window.__APP_CONFIG__ = {
  // ...其他配置
  updateServerUrl: "https://updates.customer-a.com/desktop/",
};
```

### 2.3 发布一个更新版本
1. 升级 `desktop/package.json` 的 `version`（如 `1.0.0` → `1.0.1`）。
2. 执行 `pnpm build`（或 `build:win` 等平台命令）。
3. 把 `desktop/release/` 下生成的：
   - 安装包（如 `GoWind Admin Setup 1.0.1.exe`）
   - `latest.yml`（ electron-updater 的版本元数据）
   - 差量更新块（`.blockmap`）
   
   全部上传到更新源的对应目录（保持目录结构）。
4. 已安装的客户端下次启动时会自动检测 `latest.yml`，提示升级。

### 2.4 切换/停用
- 切换源：改 `app-config.js` 的 `updateServerUrl`，或改构建环境变量重新打包。
- 停用：删除 `updateServerUrl` 配置，客户端将不检查更新。

---

## 三、CI 平台（D-6）

### 作用
自动化三平台（Win/mac/Linux）构建与发布，不在本机交叉打包。

### 可配置点
CI 配置文件按平台提供模板，选其一放入工程即可。**敏感凭据全部走 CI Secrets**，不入库。

### 3.1 GitHub Actions（模板路径：`.github/workflows/desktop-build.yml`）
- 三平台矩阵：`strategy.matrix.os: [windows-latest, macos-latest, ubuntu-latest]`
- 在仓库 **Settings → Secrets** 注入：`CSC_LINK`、`CSC_KEY_PASSWORD`、`APPLE_ID`、`APPLE_APP_SPECIFIC_PASSWORD`、`APPLE_TEAM_ID`、`UPDATE_SERVER_URL` 等。
- workflow 里用 `${{ secrets.XXX }}` 引用。

### 3.2 Jenkins
- 用 `agent` 标签指定三个平台的节点（Win/mac/Linux 各一个）。
- 凭据存入 **Jenkins Credentials**，用 `withCredentials([...])` 注入为环境变量。
- 触发器：手动 / tag 触发。

### 3.3 GitLab CI
- 用 `tags` 指定各平台 runner。
- 凭据存入 **Settings → CI/CD → Variables**（设为 Masked/Protected）。

### 3.4 统一原则
- 仅在打 tag（如 `v1.0.1`）时触发完整三平台构建发布；日常提交不跑（耗时）。
- 产物上传到更新源（见第二章）+ 作为 Release 附件。

> 实际接入时，参照上方对应平台的官方 electron-builder CI 文档；本手册提供接入要点。

---

## 四、外设能力启用（D-7）

### 作用
按业务优先级逐项启用本地能力（打印、串口、USB、读卡器等）。能力以**模块**形式组织，默认不启用，按需开启。

### 可配置点
每项能力 = 三个文件 + 一处注册：
1. 主进程 handler：`electron/main/ipc/<能力>.ts`
2. preload 暴露：`electron/preload/index.ts` 加白名单方法
3. 前端类型：`react/src` 的 `window.desktop` 类型声明（可选）
4. main 入口注册 handler

### 4.1 启用一个新能力（以"打印"为例）

**① 安装依赖**（在 `desktop/` 下）
```bash
pnpm add electron-to-printer   # 或选定打印库
```

**② 写主进程 handler**（`electron/main/ipc/print.ts`）
```ts
import { ipcMain } from 'electron';
import { print } from '../services/printer';

export function registerPrintIpc() {
  ipcMain.handle('print:direct', async (_e, opts) => {
    return print(opts);
  });
}
```

**③ 在 main 入口注册**（`electron/main/index.ts`）
```ts
import { registerPrintIpc } from './ipc/print';
app.whenReady().then(() => {
  registerPrintIpc();
  createWindow();
});
```

**④ preload 暴露**（`electron/preload/index.ts`）
```ts
const desktop = {
  isDesktop: true as const,
  ping: () => ipcRenderer.invoke('desktop:ping'),
  print: (opts?: PrintOpts) => ipcRenderer.invoke('print:direct', opts),  // 新增
};
```

**⑤ 前端调用**（react 业务代码）
```ts
const ok = await window.desktop?.print({ printerName: 'XP-80' });
```

### 4.2 能力清单与推荐库
| 能力 | 推荐库 | 跨平台注意 |
|---|---|---|
| 打印（通用） | `electron` 内置 `webContents.print` | 全平台 |
| 热敏/小票打印 | `electron-to-printer` / `node-printer` | 需各平台编译 |
| 串口 | `serialport` | 需 `@electron/rebuild` |
| USB | `usb` / `node-usb` | 需各平台编译 |
| 本地数据库 | `better-sqlite3` | prebuilt 较全 |
| 身份证/读卡器 | 厂商 SDK | 通常仅 Windows |

### 4.3 启用/禁用切换
- **禁用某能力**：注释掉 main 入口的 `registerXIpc()` + preload 对应行即可，前端调用会返回 undefined（自动降级）。
- **B/S 降级**：前端始终用 `window.desktop?.xxx` 可选链，B/S 下自动跳过该能力，无需条件编译。

---

## 附：决策状态总览

| 项 | 状态 | 配置位置 |
|---|---|---|
| D-4 代码签名 | 可选，默认不签名 | 构建时环境变量 / `builder.yml` |
| D-5 自动更新源 | 可选，默认不启用 | `app-config.js` 的 `updateServerUrl` / 构建环境变量 |
| D-6 CI 平台 | 可选，按团队选 | `.github/workflows/` 或 Jenkinsfile 或 `.gitlab-ci.yml` |
| D-7 外设能力 | 按需启用 | `electron/main/ipc/*.ts` + preload |
