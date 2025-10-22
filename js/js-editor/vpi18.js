import { HEIGHT, WIDTH, DEPTH } from './palette.js';

// Packs an array of entries { index: 0..4095, paletteIndex: 0..63 } into a Uint8Array VPI18 bitstream.
export function encodeVPI18(entries) {
  if (!Array.isArray(entries) || entries.length === 0) return new Uint8Array(0);
  const totalBits = entries.length * 18;
  const totalBytes = Math.ceil(totalBits / 8);
  const out = new Uint8Array(totalBytes);
  let bitPos = 0; // next bit to write [0..)

  for (const e of entries) {
    const idx = e.index & 0xFFF; // 12 bits
    const col = e.paletteIndex & 0x3F; // 6 bits
    const value = (idx << 6) | col; // 18 bits
    // Write 18 bits MSB->LSB relative to this value
    for (let b = 17; b >= 0; b--) {
      const bit = (value >> b) & 1;
      const byteIndex = bitPos >> 3;
      const bitIndex = 7 - (bitPos & 7);
      out[byteIndex] |= bit << bitIndex;
      bitPos++;
    }
  }
  return out;
}

// Decodes a VPI18 bitstream (Uint8Array, ArrayBuffer, or DataView) into an array of entries.
export function decodeVPI18(data) {
  const bytes = data instanceof Uint8Array ? data : new Uint8Array(data.buffer ?? data);
  const totalBits = bytes.length * 8;
  const entryCount = Math.floor(totalBits / 18);
  const entries = new Array(entryCount);
  let bitPos = 0;
  for (let i = 0; i < entryCount; i++) {
    let value = 0;
    for (let b = 0; b < 18; b++) {
      const byteIndex = bitPos >> 3;
      const bitIndex = 7 - (bitPos & 7);
      const bit = (bytes[byteIndex] >> bitIndex) & 1;
      value = (value << 1) | bit;
      bitPos++;
    }
    const idx = (value >> 6) & 0xFFF;
    const paletteIndex = value & 0x3F;
    entries[i] = { index: idx, paletteIndex };
  }
  return entries;
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
