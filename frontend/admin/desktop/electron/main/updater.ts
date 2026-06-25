import { ipcMain, dialog } from 'electron';
import { autoUpdater } from 'electron-updater';

/**
 * 自动更新 / Auto-update (electron-updater)
 *
 * 更新地址由渲染进程在启动时经 app-config.js 的 updateServerUrl 传入（D-5 解耦：运行时可改、免编译），
 * 详见 docs/caozuo/desktop-optional-config.md 第二章。未传地址 → 不检查更新（等同停用）。
 *
 * 流程：渲染进程 window.desktop.checkForUpdates(updateServerUrl)
 *      → setFeedURL(generic, url) → checkForUpdates → 自动下载 → 下载完成弹窗询问 → quitAndInstall。
 *
 * 注意：
 * - Windows 生产环境的完整自动更新依赖代码签名（D-4）；未签名时，开发期可用 dev-app-update.yml 验证流程。
 * - 开发态（!app.isPackaged）electron-updater 读取项目根的 dev-app-update.yml；生产态读 resources/app-update.yml。
 */

export interface UpdateCheckResult {
  checked: boolean;
  reason?: string;
  version?: string;
}

let wired = false;

/** 首次配置时挂载事件回调（只挂一次） */
function ensureWired(): void {
  if (wired) return;
  wired = true;
  autoUpdater.autoDownload = true; // 发现更新即后台下载
  autoUpdater.autoInstallOnAppQuit = true; // 下载完成后，应用退出时自动安装

  autoUpdater.on('update-available', (info) => {
    console.log('[updater] 发现新版本:', info?.version);
  });
  autoUpdater.on('update-not-available', () => {
    console.log('[updater] 已是最新版本');
  });
  autoUpdater.on('error', (err) => {
    console.error('[updater] 错误:', err);
  });
  autoUpdater.on('update-downloaded', (info) => {
    // 下载完成，询问是否立即重启安装（不强制）
    dialog
      .showMessageBox({
        type: 'info',
        title: '更新已就绪',
        message: `新版本 ${info?.version} 已下载完成。`,
        detail: '是否立即重启并安装？',
        buttons: ['稍后', '立即重启安装'],
        defaultId: 1,
        cancelId: 0,
      })
      .then(({ response }) => {
        if (response === 1) autoUpdater.quitAndInstall();
      });
  });
}

/** 注册自动更新相关 IPC（在 app.whenReady 中调用） */
export function initUpdater(): void {
  // 渲染进程传入运行时更新地址并触发检查
  ipcMain.handle('update:check', async (_event, url?: string): Promise<UpdateCheckResult> => {
    if (!url || typeof url !== 'string') {
      return { checked: false, reason: 'no-update-url' };
    }
    ensureWired();
    autoUpdater.setFeedURL({ provider: 'generic', url });
    try {
      const result = await autoUpdater.checkForUpdates();
      return { checked: true, version: result?.updateInfo?.version };
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      console.error('[updater] 检查更新失败:', msg);
      return { checked: false, reason: msg };
    }
  });

  // 立即重启并安装已下载的更新
  ipcMain.handle('update:install', () => {
    autoUpdater.quitAndInstall();
  });
}
