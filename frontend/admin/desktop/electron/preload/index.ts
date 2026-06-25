import { contextBridge, ipcRenderer } from 'electron';

/**
 * 桌面端能力桥接 / Desktop capability bridge
 *
 * 通过 contextBridge 向渲染进程暴露**受限白名单** API（绝不暴露 ipcRenderer / require 原始对象）。
 * B/S 环境下 window.desktop 不存在；前端用 window.desktop?.xxx 可选链调用，自动降级。
 *
 * 渲染进程侧的类型契约见 react 工程的 src/types/desktop.d.ts（与本文件导出的 DesktopBridge 保持一致）。
 * 具体能力（打印/串口/USB 等）的启用见 docs/caozuo/desktop-optional-config.md 第四章。
 */

/** 保存对话框选项（与渲染进程侧 DesktopSaveDialogOptions 契约一致） */
interface SaveDialogOptions {
  /** 建议的默认文件名，如 'export.csv' */
  defaultName?: string;
  /** 默认目录（绝对路径）；若提供则优先于 defaultName */
  defaultPath?: string;
  /** 文件类型过滤器，如 [{ name: 'CSV', extensions: ['csv'] }] */
  filters?: Array<{ name: string; extensions: string[] }>;
}

const desktop = {
  /** 桌面环境标记 — 前端据此切换 hash 路由等运行时行为 */
  isDesktop: true as const,
  /** 链路自检：返回 'pong'（P0） */
  ping: (): Promise<string> => ipcRenderer.invoke('desktop:ping'),
  /** 弹出保存对话框，返回用户选择的文件绝对路径；用户取消则返回空字符串（P2） */
  saveDialog: (opts?: SaveDialogOptions): Promise<string> =>
    ipcRenderer.invoke('dialog:save', opts ?? {}),
  /** 将二进制数据写入指定路径，覆盖已存在文件（P2） */
  writeFile: (filePath: string, data: ArrayBuffer | Uint8Array): Promise<void> =>
    ipcRenderer.invoke('file:write', { path: filePath, data }),
  /**
   * 检查更新（P4）：传入运行时更新源地址（app-config.js 的 updateServerUrl）。
   * 发现新版本会自动后台下载，下载完成后弹窗询问是否重启安装。
   */
  checkForUpdates: (url?: string): Promise<{ checked: boolean; reason?: string; version?: string }> =>
    ipcRenderer.invoke('update:check', url),
  /** 立即重启并安装已下载的更新（P4） */
  quitAndInstall: (): Promise<void> => ipcRenderer.invoke('update:install'),
  /** 打印当前窗口页面（P5 通用打印样板）；返回是否成功提交打印任务 */
  print: (opts?: { silent?: boolean; printBackground?: boolean; deviceName?: string }): Promise<boolean> =>
    ipcRenderer.invoke('print', opts ?? {}),
};

contextBridge.exposeInMainWorld('desktop', desktop);

export type DesktopBridge = typeof desktop;
