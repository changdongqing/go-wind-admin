import { ipcMain } from 'electron';

/**
 * 桌面元能力 IPC / Desktop meta-capability IPC
 *
 * P0 链路自检：渲染进程调用 window.desktop.ping() → 经 preload → 此处返回 'pong'，
 * 用于验证「渲染进程 → preload → 主进程」整条 IPC 链路是否畅通。
 */
export function registerDesktopIpc(): void {
  ipcMain.handle('desktop:ping', () => 'pong');
}
