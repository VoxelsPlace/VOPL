import { HEIGHT, WIDTH, DEPTH } from './palette.js';

let examplesIndex = new Map();
let examplesList = [];

function normalizeLayerRows(rows) {
  const out = new Array(DEPTH);
  for (let z = 0; z < DEPTH; z++) {
    let row = rows[z];
    if (typeof row !== 'string') row = '0'.repeat(WIDTH);
    const tokenized = row.indexOf(',') !== -1 || /\s/.test(row);
    if (tokenized) {
      out[z] = row.trim();
    } else {
      let s = row;
      if (s.length < WIDTH) s = s + '0'.repeat(WIDTH - s.length);
      if (s.length > WIDTH) s = s.slice(0, WIDTH);
      out[z] = s;
    }
  }
  return out;
}

export async function loadExamples() {
  const res = await fetch('./examples.json');
  const json = await res.json();
  examplesIndex.clear();
  examplesList = [];
  const list = json.examples || [];
  for (const ex of list) {
    const srcLayers = ex.layers || [];
    const layers = new Array(HEIGHT);
    for (let y = 0; y < HEIGHT; y++) {
      const entry = Array.isArray(srcLayers[y]) ? srcLayers[y] : [];
      layers[y] = normalizeLayerRows(entry);
    }
    const item = { id: ex.id, name: ex.name, layers };
    examplesIndex.set(ex.id, item);
    examplesList.push({ id: ex.id, name: ex.name });
  }
}

export function getExampleById(id) {
  return examplesIndex.get(id) || null;
}

export function getExamples() {
  return examplesList.slice();
}

