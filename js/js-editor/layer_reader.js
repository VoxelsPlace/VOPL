import { HEIGHT, WIDTH, DEPTH } from './palette.js';
import { updateVoxel } from './input.js';

export function applyLayers(layers) {
  for (let y = 0; y < HEIGHT; y++) {
    const rows = layers[y];
    for (let z = 0; z < DEPTH; z++) {
      const row = rows[z];
      const values = parseRow(row);
      for (let x = 0; x < WIDTH; x++) {
        const v = values[x] | 0;
        if (v !== 0) updateVoxel(x, y, z, v);
      }
    }
  }
}

function parseRow(row) {
  if (row.indexOf(',') !== -1 || /\s/.test(row)) {
    const tokens = row.split(/[,\s]+/).filter(Boolean).map(t => {
      const n = parseInt(t, 10);
      if (!Number.isFinite(n) || n < 0 || n > 63) return 0;
      return n;
    });
    if (tokens.length < WIDTH) {
      const out = tokens.slice();
      while (out.length < WIDTH) out.push(0);
      return out;
    }
    return tokens.slice(0, WIDTH);
  }
  const out = new Array(WIDTH);
  for (let i = 0; i < WIDTH; i++) {
    const ch = row[i] || '0';
    const v = ch >= '0' && ch <= '9' ? (ch.charCodeAt(0) - 48) : 0;
    out[i] = v;
  }
  return out;
}

