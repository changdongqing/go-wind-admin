/**
 * 桌面客户端能力桥接类型 / Desktop bridge type contract
 * ----------------------------------------------------------------------------
 * window.desktop 仅在桌面端（Electron）存在 —— 由 desktop 工程的 preload 通过
 * contextBridge 注入；B/S（浏览器）环境下为 undefined。
 *
 * 前端统一用 `window.desktop?.xxx` 可选链调用，缺失时自动降级（隐藏/禁用相关能力），
 * 因此同一份 React 产物既能跑在浏览器，也能跑在 Electron 壳里，无需构建分叉。
 *
 * 本文件是渲染进程侧的「接口契约」声明，与 desktop 工程 preload 的实现保持一致；
 * react 工程不依赖 desktop 工程（B/S 与桌面解耦，见 docs/desktop/desktop-client-plan.md 第四章）。
 * 能力清单见同文档第八章；安全基线（白名单制、禁暴露原始对象）见第九章。
 */

/** 保存对话框选项（对齐 Electron dialog.showSaveDialog 的常用参数子集） */
interface DesktopSaveDialogOptions {
    /** 默认文件名 */
    defaultName?: string;
    /** 默认所在目录（绝对路径） */
    defaultPath?: string;
    /** 文件类型过滤器，如 [{ name: 'CSV', extensions: ['csv'] }] */
    filters?: Array<{ name: string; extensions: string[] }>;
}

/** 打印选项（对齐 Electron webContents.print 的常用参数子集） */
interface DesktopPrintOptions {
    /** 静默打印：不弹打印机选择对话框，直接用默认/指定打印机 */
    silent?: boolean;
    /** 是否打印背景色/背景图 */
    printBackground?: boolean;
    /** 指定打印机名（Windows/Linux 有效；macOS 忽略） */
    deviceName?: string;
}

/** 自动更新检查结果（与 desktop 工程 updater.ts 的 UpdateCheckResult 契约一致） */
interface DesktopUpdateCheckResult {
    checked: boolean;
    reason?: string;
    version?: string;
}

/**
 * window.desktop 桥接对象（桌面端存在，B/S 下不存在）
 * 暴露的是受限白名单 API；preload 绝不暴露 ipcRenderer / require 等原始对象。
 */
interface DesktopBridge {
    /** 桌面环境标记 — 前端据此切换 hash 路由等运行时行为 */
    readonly isDesktop: true;
    /** 链路自检：返回 'pong'（P0 验证「渲染→preload→主进程」链路用） */
    ping(): Promise<string>;
    /** 弹出保存对话框，返回用户选择的文件绝对路径；用户取消则返回空字符串 */
    saveDialog(opts?: DesktopSaveDialogOptions): Promise<string>;
    /** 将二进制数据写入指定路径（覆盖已存在文件） */
    writeFile(filePath: string, data: ArrayBuffer | Uint8Array): Promise<void>;
    /** 检查更新：传入更新源地址（app-config.js 的 updateServerUrl）；发现新版本自动下载并提示 */
    checkForUpdates(url?: string): Promise<DesktopUpdateCheckResult>;
    /** 立即重启并安装已下载的更新 */
    quitAndInstall(): Promise<void>;
    /** 打印当前窗口页面；返回是否成功提交打印任务 */
    print(opts?: DesktopPrintOptions): Promise<boolean>;
}

interface Window {
    /** 桌面端桥接对象；仅在 Electron 环境存在，B/S 下为 undefined */
    desktop?: DesktopBridge;
}
