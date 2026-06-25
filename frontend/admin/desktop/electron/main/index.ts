import { app, BrowserWindow, shell } from 'electron';
import path from 'node:path';
import { registerAllIpc } from './ipc';
import { initUpdater } from './updater';

/**
 * 主进程入口 / Main process entry
 *
 * 开发态：加载 react 的 dev server（默认 http://localhost:7000）
 * 生产态：加载打包进来的 react 产物（renderer-dist/index.html）
 */
const isDev = !app.isPackaged;
const DEV_SERVER_URL = process.env.DEV_SERVER_URL ?? 'http://localhost:7000';

async function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 1024,
    minHeight: 680,
    show: false,
    autoHideMenuBar: true,
    backgroundColor: '#ffffff',
    title: 'GoWind Admin',
    webPreferences: {
      preload: path.join(__dirname, '../preload/index.js'),
      // ===== 安全基线（见 docs/desktop/desktop-client-plan.md 第九章）=====
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
      webSecurity: true,
    },
  });

  // 窗口准备好再显示，避免白屏闪烁
  win.once('ready-to-show', () => win.show());

  // 外部链接（http/https）交给系统浏览器，不在应用内打开
  win.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  if (isDev) {
    await win.loadURL(DEV_SERVER_URL);
    win.webContents.openDevTools({ mode: 'detach' });
  } else {
    await win.loadFile(path.join(__dirname, '../../renderer-dist/index.html'));
  }
}

app.whenReady().then(() => {
  // 注册全部 IPC 能力（按能力分文件，见 docs/caozuo/desktop-optional-config.md 第四章）
  registerAllIpc();
  // 自动更新（P4）：注册 update:* IPC；实际检查由渲染进程传入 updateServerUrl 后触发
  initUpdater();
  createWindow();
});

app.on('window-all-closed', () => {
  // macOS 约定：关窗后应用继续驻留 dock
  if (process.platform !== 'darwin') app.quit();
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});
