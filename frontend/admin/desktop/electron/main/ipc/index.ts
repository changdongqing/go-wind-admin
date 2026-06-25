/**
 * IPC 能力注册聚合 / IPC capability registration aggregator
 *
 * 每项本地能力一个 register 函数，在此统一注册。
 * 新增能力（打印/串口/USB 等）时：在对应 ipc/<能力>.ts 实现 registerXxxIpc，并加到下方列表。
 * 启停约定见 docs/caozuo/desktop-optional-config.md 第四章。
 */
import { registerDesktopIpc } from './desktop';
import { registerFileIpc } from './file';
import { registerPrintIpc } from './print';

export function registerAllIpc(): void {
  registerDesktopIpc(); // P0 链路自检（ping）
  registerFileIpc(); // P2 保存对话框 + 写文件
  registerPrintIpc(); // P5 通用打印（webContents.print）
}
