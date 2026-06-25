import { ipcMain, BrowserWindow } from 'electron';

/**
 * 打印能力 IPC / Print capability IPC
 *
 * 用 Electron 内置 webContents.print（全平台、零原生模块编译）实现「通用打印」，
 * 作为外设能力的第二个样板（首个为 file.ts）。热敏/小票打印机等需用专用库时，
 * 参照 docs/caozuo/desktop-optional-config.md 第四章新增 ipc/<能力>.ts 即可。
 *
 * 调用：const ok = await window.desktop?.print({ silent: true });
 * 说明：打印的是当前窗口页面内容；如需打印指定 HTML，可先在隐藏窗口加载后打印（按需扩展）。
 */

export interface PrintOpts {
  /** 静默打印：不弹打印机选择对话框，直接用默认打印机（或 deviceName 指定的打印机） */
  silent?: boolean;
  /** 是否打印背景色/背景图 */
  printBackground?: boolean;
  /** 指定打印机名（Windows/Linux 有效；macOS 忽略，走系统对话框） */
  deviceName?: string;
}

export function registerPrintIpc(): void {
  ipcMain.handle('print', async (event, opts: PrintOpts = {}): Promise<boolean> => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (!win) {
      console.error('[print] 找不到所属窗口');
      return false;
    }
    try {
      await win.webContents.print({
        silent: opts.silent,
        printBackground: opts.printBackground,
        deviceName: opts.deviceName,
      });
      return true;
    } catch (err) {
      console.error('[print] 打印失败:', err);
      return false;
    }
  });
}
