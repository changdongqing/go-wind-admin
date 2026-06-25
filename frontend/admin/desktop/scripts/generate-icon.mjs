#!/usr/bin/env node
/**
 * 生成桌面端占位图标 / Generate desktop placeholder icon (256×256 PNG-in-ICO)
 * ----------------------------------------------------------------------------
 * electron-builder 要求 Windows 安装器图标 ≥256×256；react 现有 favicon.ico 仅 32×32、
 * logo.png 仅 200×200，都不达标。本脚本用纯 Node（zlib，零依赖）生成一张 256×256 的
 * 占位图标（蓝色底 + 白色风纹，呼应「风行/GoWind」），写入 desktop/build/icon.ico。
 *
 * 这是**占位图标** —— 正式交付请用品牌 256×256+ 多分辨率图标替换 desktop/build/icon.ico
 * （build/ 目录已 gitignore，作为部署期资源，与 app-config.js 同属客户化资产）。
 *
 * 运行：`node scripts/generate-icon.mjs`，或由 sync-react-dist.mjs 在图标缺失时调用。
 */
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { fileURLToPath } from 'node:url';

const SIZE = 256;
const BG = [22, 119, 255]; // #1677FF（Ant Design 主色蓝）
const FG = [255, 255, 255]; // 白色风纹

/** 构造 SIZE×SIZE 的 RGBA 像素：蓝底 + 三条递减的横向风纹 */
function buildPixels() {
  const buf = Buffer.alloc(SIZE * SIZE * 4);
  const cx = SIZE / 2;
  for (let y = 0; y < SIZE; y++) {
    for (let x = 0; x < SIZE; x++) {
      const i = (y * SIZE + x) * 4;
      let c = BG;
      // 三条风纹（向右递减），居中偏左对齐
      const wind =
        (y >= 100 && y <= 120 && x >= cx - 72 && x <= cx + 72) || // 最长
        (y >= 124 && y <= 144 && x >= cx - 72 && x <= cx + 36) || // 中
        (y >= 148 && y <= 168 && x >= cx - 72 && x <= cx + 0); // 最短
      if (wind) c = FG;
      buf[i] = c[0];
      buf[i + 1] = c[1];
      buf[i + 2] = c[2];
      buf[i + 3] = 255;
    }
  }
  return buf;
}

/** 编码一个 PNG chunk：length + type + data + crc32(type+data) */
function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const typeBuf = Buffer.from(type, 'ascii');
  const crc = Buffer.alloc(4);
  crc.writeUInt32LE(zlib.crc32(Buffer.concat([typeBuf, data])), 0);
  return Buffer.concat([len, typeBuf, data, crc]);
}

/** RGBA → PNG（8-bit RGBA，filter=0 扫描行） */
function encodePng(rgba) {
  const sig = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(SIZE, 0);
  ihdr.writeUInt32BE(SIZE, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // color type RGBA
  ihdr[10] = 0;
  ihdr[11] = 0;
  ihdr[12] = 0;
  const rows = [];
  for (let y = 0; y < SIZE; y++) {
    rows.push(Buffer.from([0])); // filter byte
    rows.push(rgba.subarray(y * SIZE * 4, (y + 1) * SIZE * 4));
  }
  const idat = zlib.deflateSync(Buffer.concat(rows));
  return Buffer.concat([sig, chunk('IHDR', ihdr), chunk('IDAT', idat), chunk('IEND', Buffer.alloc(0))]);
}

/** 把 PNG 包成单条目的 ICO（256×256，w/h 字节为 0 表示 256） */
function pngToIco(png) {
  const dir = Buffer.alloc(6);
  dir.writeUInt16LE(0, 0); // reserved
  dir.writeUInt16LE(1, 2); // type = icon
  dir.writeUInt16LE(1, 4); // count = 1
  const entry = Buffer.alloc(16);
  entry[0] = 0; // width 256 (0 ⇒ 256)
  entry[1] = 0; // height 256
  entry[2] = 0; // color count
  entry[3] = 0; // reserved
  entry.writeUInt16LE(1, 4); // planes
  entry.writeUInt16LE(32, 6); // bits per pixel
  entry.writeUInt32LE(png.length, 8); // image size
  entry.writeUInt32LE(22, 12); // offset = 6 + 16
  return Buffer.concat([dir, entry, png]);
}

/** 生成 256×256 占位图标并写入 targetPath */
export function generateIconFile(targetPath) {
  const png = encodePng(buildPixels());
  const ico = pngToIco(png);
  fs.mkdirSync(path.dirname(targetPath), { recursive: true });
  fs.writeFileSync(targetPath, ico);
  return targetPath;
}

// 直接执行：生成 desktop/build/icon.ico
const __filename = fileURLToPath(import.meta.url);
if (path.resolve(process.argv[1] ?? '') === __filename) {
  const target = path.resolve(path.dirname(__filename), '../build/icon.ico');
  generateIconFile(target);
  console.log(`[icon] 已生成占位图标 ${target}（256×256，可替换为正式品牌图标）`);
}
