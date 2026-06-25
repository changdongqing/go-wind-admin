import { ipcMain, dialog, BrowserWindow } from 'electron';
import fs from 'node:fs/promises';

/**
 * 文件能力 IPC / File capability IPC
 *
 * 提供「保存对话框 + 写文件」样板（P2），是其余本地能力（打印/串口/外设）的参照模板。
 * 渲染进程经 preload 暴露的 window.desktop.saveDialog / writeFile 调用：
 *   const p = await window.desktop?.saveDialog({ defaultName: 'export.csv' });
 *   if (p) await window.desktop?.writeFile(p, bytes);
 *
 * 安全：仅在主进程（Node 环境）接触文件系统；渲染进程只通过白名单 IPC 触达，无 fs 句柄泄漏。
 */

/** 保存对话框选项（与渲染进程侧 DesktopSaveDialogOptions 契约一致） */
interface SaveDialogOpts {
  /** 建议的默认文件名，如 'export.csv' */
  defaultName?: string;
  /** 默认目录（绝对路径）；若提供则优先于 defaultName */
  defaultPath?: string;
  /** 文件类型过滤器 */
  filters?: Array<{ name: string; extensions: string[] }>;
}

/** 写文件 IPC 的载荷结构 */
interface WriteFilePayload {
  path: string;
  data: ArrayBuffer | Uint8Array;
}

/** 把渲染进程传来的二进制统一成 Node Buffer（兼容 ArrayBuffer 与 Uint8Array 视图） */
function toBuffer(data: ArrayBuffer | Uint8Array): Buffer {
  if (data instanceof ArrayBuffer) return Buffer.from(data);
  // Uint8Array：按视图边界截取底层 buffer，避免共享 buffer 越界读取
  return Buffer.from(data.buffer, data.byteOffset, data.byteLength);
}

export function registerFileIpc(): void {
  // 保存对话框：返回用户选择的文件绝对路径；用户取消则返回空字符串
  ipcMain.handle('dialog:save', async (event, opts: SaveDialogOpts = {}) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    // Electron 的 SaveDialogOptions 无 defaultName，文件名建议通过 defaultPath 传达
    const dialogOpts = {
      defaultPath: opts.defaultPath ?? opts.defaultName,
      filters: opts.filters ?? [{ name: 'All Files', extensions: ['*'] }],
    };
    const result = win
      ? await dialog.showSaveDialog(win, dialogOpts)
      : await dialog.showSaveDialog(dialogOpts);
    return result.canceled ? '' : (result.filePath ?? '');
  });

  // 写文件：将二进制数据写入指定路径（覆盖已存在文件）
  ipcMain.handle('file:write', async (_event, payload: WriteFilePayload) => {
    await fs.writeFile(payload.path, toBuffer(payload.data));
  });
}
