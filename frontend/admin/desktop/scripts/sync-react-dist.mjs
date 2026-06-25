#!/usr/bin/env node
/**
 * 同步 react 构建产物到 desktop 的打包输入目录 / Sync react build output into desktop's package source.
 * ----------------------------------------------------------------------------
 * 解耦支点（见 docs/desktop/desktop-client-plan.md 第四章）：
 *   react 工程是生产者（产出标准 Web 产物 dist/），desktop 工程是消费者（把 dist + Electron 壳 → 安装包）。
 *   本脚本由 desktop 的 build 流程自动调用，react 工程本身不感知桌面端。
 *
 * 做三件事：
 *   1. 把 ../react/dist 整体复制到 ./renderer-dist（electron-builder 的打包内容之一）。
 *   2. 把 renderer-dist/index.html 里指向 public 资源的「根绝对路径」改写为相对路径
 *      （/app-config.js、/favicon.ico 等）—— Electron 用 file:// 加载时绝对路径会指向磁盘根，
 *      必须改成相对路径才能解析。react 源码与 B/S 产物【不动】（B/S 仍用绝对路径，零影响）。
 *   3. 若 desktop/build/icon.ico 不存在，生成一张 256×256 占位图标（见 generate-icon.mjs），
 *      满足 electron-builder「≥256×256」要求；不覆盖已存在的自定义图标。
 *
 * 运行：`node scripts/sync-react-dist.mjs`（由 `pnpm sync:react` 调用）。
 * 前置：react 已以相对 base 构建 —— `pnpm --dir ../react run build -- --base=./`（由 `pnpm build:react` 调用）。
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { generateIconFile } from './generate-icon.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const desktopRoot = path.resolve(__dirname, '..');
const reactRoot = path.resolve(desktopRoot, '../react');
const reactDist = path.join(reactRoot, 'dist');
const rendererDist = path.join(desktopRoot, 'renderer-dist');
const buildDir = path.join(desktopRoot, 'build');

/** 递归复制目录（覆盖目标） */
function copyDirSync(src, dest) {
  fs.rmSync(dest, { recursive: true, force: true });
  fs.mkdirSync(dest, { recursive: true });
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const s = path.join(src, entry.name);
    const d = path.join(dest, entry.name);
    if (entry.isDirectory()) copyDirSync(s, d);
    else fs.copyFileSync(s, d);
  }
}

/**
 * 把 index.html 中 public 资源的根绝对引用改写为相对引用：
 *   src="/x"  → src="./x"   href="/x" → href="./x"
 * 仅匹配根绝对路径（以 "/" 开头），不影响已是相对路径的打包产物引用（base: './' 下本就是 ./assets/...）。
 */
function relativizeHtmlRootPaths(htmlPath) {
  if (!fs.existsSync(htmlPath)) return;
  let html = fs.readFileSync(htmlPath, 'utf8');
  const before = html;
  html = html.replace(/(src|href)\s*=\s*"\/(?!\/)/g, '$1="./');
  // 单引号变体
  html = html.replace(/(src|href)\s*=\s*'\/(?!\/)/g, "$1='./");
  if (html !== before) {
    fs.writeFileSync(htmlPath, html);
    console.log('[sync] 已将 index.html 的根绝对路径改写为相对路径（适配 file:// 加载）');
  }
}

/** 生成 Windows 安装器图标（仅当缺失时；不覆盖用户自定义图标） */
function seedBuildIcon() {
  const target = path.join(buildDir, 'icon.ico');
  if (fs.existsSync(target)) return; // 不覆盖用户自定义图标
  generateIconFile(target);
  console.log('[sync] 已生成 build/icon.ico（256×256 占位图标，可替换为正式品牌图标）');
}

function main() {
  if (!fs.existsSync(reactDist)) {
    console.error(`[sync] ✗ 未找到 react 构建产物：${reactDist}`);
    console.error('[sync]   请先执行 `pnpm build:react`（或由 `pnpm build` 自动触发）。');
    process.exit(1);
  }
  if (!fs.existsSync(path.join(reactDist, 'index.html'))) {
    console.error(`[sync] ✗ react 构建产物缺少 index.html：${reactDist}`);
    console.error('[sync]   请确认 react 已成功执行 `vite build`。');
    process.exit(1);
  }

  console.log(`[sync] 复制 ${path.relative(desktopRoot, reactDist)} → ${path.relative(desktopRoot, rendererDist)}`);
  copyDirSync(reactDist, rendererDist);
  console.log('[sync] renderer-dist 就绪');

  relativizeHtmlRootPaths(path.join(rendererDist, 'index.html'));
  seedBuildIcon();

  console.log('[sync] ✓ 完成');
}

main();
