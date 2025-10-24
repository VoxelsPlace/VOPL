import { HEIGHT, WIDTH, DEPTH } from './palette.js';

// Converts linear index -> x,y,z with index = x + y*W + z*W*H
export function indexToXYZ(index) {
  const w = WIDTH, h = HEIGHT, d = DEPTH;
  const wh = w * h;
  const z = Math.floor(index / wh);
  const rem = index - z * wh;
  const y = Math.floor(rem / w);
  const x = rem - y * w;
  return { x, y, z };
}

export function xyzToIndex(x, y, z) {
  return x + y * WIDTH + z * WIDTH * HEIGHT;
}

// Build sparse entries from voxel grid: [{ index, color }]
export function entriesFromVoxelGrid(voxelGrid) {
  const out = [];
  for (let y = 0; y < HEIGHT; y++) {
    for (let z = 0; z < DEPTH; z++) {
      for (let x = 0; x < WIDTH; x++) {
        const val = voxelGrid[y][x][z] | 0;
        if (val !== 0) out.push({ index: xyzToIndex(x, y, z), color: val & 0xFF });
      }
    }
  }
  return out;
}

// Encode entries into updates JSON: { "<chunkId>": { "<index>": color } }
export function encodeUpdatesJSON(entries, chunkId = '0') {
  const inner = {};
  for (const e of entries) {
    inner["" + (e.index | 0)] = e.color | 0;
  }
  return JSON.stringify({ [chunkId]: inner }, null, 2);
}

// Apply updates JSON (string or object) using provided setter
export function applyUpdatesJSON(jsonInput, setVoxelFn) {
  const obj = typeof jsonInput === 'string' ? JSON.parse(jsonInput) : jsonInput;
  const firstChunk = Object.keys(obj)[0];
  if (!firstChunk) return;
  const indices = obj[firstChunk] || {};
  for (const k in indices) {
    const idx = (k|0);
    const col = indices[k] | 0;
    const { x, y, z } = indexToXYZ(idx);
    if (x < 0 || x >= WIDTH || y < 0 || y >= HEIGHT || z < 0 || z >= DEPTH) continue;
    setVoxelFn(x, y, z, col);
  }
}
