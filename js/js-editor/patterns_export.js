import { HEIGHT, WIDTH, DEPTH, usableColors } from './palette.js';
import { voxelGrid } from './state.js';
import { updateVoxel } from './input.js';
import { applyLayers } from './layer_reader.js';
import { getExampleById } from './examples_loader.js';
import { entriesFromVoxelGrid, encodeUpdatesJSON } from './updates_json.js';

export function clearAllVoxels() {
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        if (voxelGrid[y][x][z] !== 0) updateVoxel(x, y, z, 0);
      }
    }
  }
}

const fract = (x) => x - Math.floor(x);
function hash3(ix, iy, iz, seed) { return fract(Math.sin(ix * 127.1 + iy * 311.7 + iz * 74.7 + seed) * 43758.5453); }
const smoothstep01 = (t) => t * t * (3 - 2 * t);
function valueNoise3(x, y, z, seed) {
  const xi = Math.floor(x), yi = Math.floor(y), zi = Math.floor(z);
  const tx = x - xi, ty = y - yi, tz = z - zi;
  function h(i, j, k) { return hash3(i, j, k, seed); }
  const c000 = h(xi, yi, zi), c100 = h(xi + 1, yi, zi), c010 = h(xi, yi + 1, zi), c110 = h(xi + 1, yi + 1, zi), c001 = h(xi, yi, zi + 1), c101 = h(xi + 1, yi, zi + 1), c011 = h(xi, yi + 1, zi + 1), c111 = h(xi + 1, yi + 1, zi + 1);
  const sx = smoothstep01(tx), sy = smoothstep01(ty), sz = smoothstep01(tz);
  const nx00 = c000 * (1 - sx) + c100 * sx;
  const nx10 = c010 * (1 - sx) + c110 * sx;
  const nx01 = c001 * (1 - sx) + c101 * sx;
  const nx11 = c011 * (1 - sx) + c111 * sx;
  const nxy0 = nx00 * (1 - sy) + nx10 * sy;
  const nxy1 = nx01 * (1 - sy) + nx11 * sy;
  return nxy0 * (1 - sz) + nxy1 * sz;
}
function fbm3(x, y, z, seed, octaves = 3) {
  let val = 0, amp = 0.5, freq = 1;
  for (let i = 0; i < octaves; i++) {
    val += valueNoise3(x * freq, y * freq, z * freq, seed + i * 13.37) * amp;
    amp *= 0.5; freq *= 2;
  }
  return val;
}

export function randomNoise() {
  clearAllVoxels();
  const seed = Math.random() * 10000;
  const scaleX = 0.35, scaleY = 0.55, scaleZ = 0.35;
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        const n = fbm3(x * scaleX, y * scaleY, z * scaleZ, seed, 3);
        const threshold = 0.58 - y * 0.06;
        if (n > threshold) {
          const idx = (Math.floor(n * 997 + x * 17 + y * 31 + z * 13) % usableColors.length + usableColors.length) % usableColors.length;
          updateVoxel(x, y, z, usableColors[idx]);
        } else {
          updateVoxel(x, y, z, 0);
        }
      }
    }
  }
}

export function fillChunk() {
  clearAllVoxels();
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        const colorIdx = usableColors[Math.floor(Math.random() * usableColors.length)];
        updateVoxel(x, y, z, colorIdx);
      }
    }
  }
}

export function genRainbow() {
  clearAllVoxels();
  const maxDist = Math.sqrt(WIDTH**2 + HEIGHT**2 + DEPTH**2);
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        const dist = Math.sqrt(x**2 + y**2 + z**2);
        const normalized = dist / maxDist;
        const colorIndex = Math.floor(normalized * usableColors.length);
        const color = usableColors[Math.min(colorIndex, usableColors.length - 1)];
        updateVoxel(x, y, z, color);
      }
    }
  }
}

export function genStripes() {
  clearAllVoxels();
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        const colorIndex = z % usableColors.length;
        updateVoxel(x, y, z, usableColors[colorIndex]);
      }
    }
  }
}

export function genSphere() {
  clearAllVoxels();
  const cX = WIDTH / 2 - 0.5, cY = HEIGHT / 2 - 0.5, cZ = DEPTH / 2 - 0.5;
  const radius = Math.min(WIDTH, HEIGHT, DEPTH) / 2;
  for (let y = 0; y < HEIGHT; y++) {
    for (let x = 0; x < WIDTH; x++) {
      for (let z = 0; z < DEPTH; z++) {
        const dist = Math.sqrt((x-cX)**2 + (y-cY)**2 + (z-cZ)**2);
        if (dist <= radius) {
          const normalized = dist / radius;
          const colorIndex = Math.floor(normalized * usableColors.length);
          const color = usableColors[Math.min(colorIndex, usableColors.length - 1)];
          updateVoxel(x, y, z, color);
        }
      }
    }
  }
}

// Generates exactly 3 layers, each containing a 2x2 contour (a 2x2 square),
// using random colors for each layer. All other voxels are cleared.
export function gen2x2Contour3Layers() {
  clearAllVoxels();
  // We interpret "2x2 width height" as a border thickness of 2 voxels on a large square,
  // leaving a 1-voxel margin around the edges,
  // with ONLY 2 blocks height (two contiguous layers). Colors are noise-based per voxel.
  const thickness = 2;
  const margin = 1;
  const x0 = margin;
  const x1 = Math.max(margin, WIDTH - 1 - margin); // inclusive
  const z0 = margin;
  const z1 = Math.max(margin, DEPTH - 1 - margin); // inclusive

  // Ensure we have at least room for thickness on both sides
  if (x1 - x0 + 1 < thickness * 2 || z1 - z0 + 1 < thickness * 2) return;

  // Use exactly the bottom two layers when available: y=0 and y=1
  const yLayers = HEIGHT >= 2 ? [0, 1] : [0];

  // Seed and scale for coherent noise
  const seed = Math.random() * 10000;
  const scale = 0.35;

  for (const y of yLayers) {
    for (let z = z0; z <= z1; z++) {
      for (let x = x0; x <= x1; x++) {
        const onLeft = x - x0 < thickness;
        const onRight = x1 - x < thickness;
        const onTop = z - z0 < thickness;
        const onBottom = z1 - z < thickness;
        const onBorder = onLeft || onRight || onTop || onBottom;
        if (!onBorder) continue;
        const n = fbm3(x * scale, y * scale, z * scale, seed, 2);
        const idx = Math.floor(n * usableColors.length) % usableColors.length;
        const color = usableColors[idx];
        updateVoxel(x, y, z, color);
      }
    }
  }
}

// Build an Updates JSON string from the current voxel grid (single chunk id '0')
export function buildUpdatesJSON() {
  const entries = entriesFromVoxelGrid(voxelGrid);
  return encodeUpdatesJSON(entries, '0');
}

// Trigger a client-side download of the Updates JSON
export function downloadUpdatesJSON() {
  const json = buildUpdatesJSON();
  const outName = (document.getElementById('outputName').value || 'updates.json').trim();
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = outName.endsWith('.json') ? outName : `${outName}`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export function genSquare() {
  const ex = getExampleById('house_square_black');
  if (!ex) return;
  clearAllVoxels();
  applyLayers(ex.layers);
}

// Optional: compute Updates JSON diff between two voxel grids (A->B) returning added/changed voxels only
export function buildUpdatesJSONDiff(prevGrid, nextGrid) {
  const entries = [];
  for (let y = 0; y < HEIGHT; y++) {
    for (let z = 0; z < DEPTH; z++) {
      for (let x = 0; x < WIDTH; x++) {
        const a = prevGrid[y][x][z] | 0;
        const b = nextGrid[y][x][z] | 0;
        if (a !== b && b !== 0) {
          const index = x + y * WIDTH + z * WIDTH * HEIGHT;
          entries.push({ index, color: b & 0xFF });
        }
      }
    }
  }
  return encodeUpdatesJSON(entries, '0');
}

