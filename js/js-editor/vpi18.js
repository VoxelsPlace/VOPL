import { HEIGHT, WIDTH, DEPTH } from './palette.js';
import { wasmApi } from './wasm_api.js';

// Packs an array of entries { index: 0..4095, paletteIndex: 0..63 } into a Uint8Array VPI18 bitstream.
export function encodeVPI18(entries) {
  if (!Array.isArray(entries) || entries.length === 0) return new Uint8Array(0);
  const n = entries.length;
  const indices = new Uint16Array(n);
  const colors = new Uint8Array(n);
  for (let i = 0; i < n; i++) {
    const e = entries[i];
    indices[i] = e.index & 0x0FFF;
    colors[i] = e.paletteIndex & 0x3F;
  }
  return wasmApi.vpiEncodeEntries(indices, colors);
}

// Decodes a VPI18 bitstream (Uint8Array, ArrayBuffer, or DataView) into an array of entries.
export function decodeVPI18(data) {
  const bytes = data instanceof Uint8Array ? data : new Uint8Array(data.buffer ?? data);
  const res = wasmApi.vpiDecodeEntries(bytes);
  if (typeof res === 'string') throw new Error(res);
  const idx = res.indices, col = res.colors;
  const out = new Array(idx.length);
  for (let i = 0; i < idx.length; i++) out[i] = { index: idx[i], paletteIndex: col[i] };
  return out;
}

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

// Apply a VPI18 bitstream directly to the scene using a provided setter (updateVoxel)
export function applyVPI18(bitstream, setVoxelFn) {
  const entries = decodeVPI18(bitstream);
  for (const e of entries) {
    const { x, y, z } = indexToXYZ(e.index);
    if (x < 0 || x >= WIDTH || y < 0 || y >= HEIGHT || z < 0 || z >= DEPTH) continue;
    setVoxelFn(x, y, z, e.paletteIndex | 0); // 0 clears, non-zero sets
  }
}

// Builds active voxel entries from the current grid (voxelGrid[y][x][z])
export function entriesFromVoxelGrid(voxelGrid) {
  const out = [];
  for (let y = 0; y < HEIGHT; y++) {
    for (let z = 0; z < DEPTH; z++) {
      for (let x = 0; x < WIDTH; x++) {
        const val = voxelGrid[y][x][z] | 0;
        if (val !== 0) out.push({ index: xyzToIndex(x, y, z), paletteIndex: val & 0x3F });
      }
    }
  }
  return out;
}

// Helper for examples/tests: create a small demo VPI18 diagonal at y=0
export function buildDemoVPI18(paletteIndex = 38) {
  const entries = [];
  const idx = (x, y, z) => x + y * WIDTH + z * WIDTH * HEIGHT;
  for (let i = 0; i < Math.min(16, Math.min(WIDTH, DEPTH)); i++) {
    entries.push({ index: idx(i, 0, i), paletteIndex });
  }
  return encodeVPI18(entries);
}
