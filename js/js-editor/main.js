import { renderer, camera, animate, onResize } from './scene.js';
import { setTool } from './state.js';
import { populateColorPalette, populateExampleButtons } from './ui.js';
import { onPointerMove, onPointerDown } from './input.js';
import { clearAllVoxels, randomNoise, fillChunk, genRainbow, genStripes, genSphere, gen2x2Contour3Layers, downloadVPI18 } from './patterns_export.js';
import { applyVPI18, buildDemoVPI18 } from './vpi18.js';
import { updateVoxel } from './input.js';
import { loadExamples } from './examples_loader.js';
import { applyLayers } from './layer_reader.js';
import { getExampleById } from './examples_loader.js';

window.addEventListener('resize', onResize);
window.addEventListener('pointermove', (e) => onPointerMove(e, renderer, camera));
window.addEventListener('pointerdown', (e) => onPointerDown(e, renderer, camera));
window.addEventListener('contextmenu', e => e.preventDefault());

document.getElementById('paintTool').addEventListener('click', () => { document.body.dataset.tool = 'paint'; setTool('paint'); });
document.getElementById('eraseTool').addEventListener('click', () => { document.body.dataset.tool = 'erase'; setTool('erase'); });

document.getElementById('clearChunk').addEventListener('click', clearAllVoxels);
document.getElementById('randomNoise').addEventListener('click', randomNoise);
document.getElementById('fillChunk').addEventListener('click', fillChunk);
document.getElementById('genRainbow').addEventListener('click', genRainbow);
document.getElementById('genStripes').addEventListener('click', genStripes);
document.getElementById('genSphere').addEventListener('click', genSphere);
document.getElementById('gen2x2Contour3Layers').addEventListener('click', gen2x2Contour3Layers);

document.getElementById('exportVPI18').addEventListener('click', downloadVPI18);
document.getElementById('importVPI18Btn').addEventListener('click', () => document.getElementById('importVPI18').click());
document.getElementById('importVPI18').addEventListener('change', async (e) => {
  const file = e.target.files?.[0];
  if (!file) return;
  const buf = new Uint8Array(await file.arrayBuffer());
  // Clear first
  clearAllVoxels();
  applyVPI18(buf, updateVoxel);
  e.target.value = '';
});

await loadExamples();
const examplesSection = document.querySelector('#controls .ui-group:nth-of-type(2) .generator-grid');
const parent = examplesSection.parentElement;
const examplesButtons = populateExampleButtons((id) => {
  const ex = getExampleById(id);
  if (!ex) return;
  clearAllVoxels();
  applyLayers(ex.layers);
});
parent.insertBefore(examplesButtons, examplesSection.nextSibling);
populateColorPalette();
document.body.dataset.tool = 'paint';
setTool('paint');
animate();

// Optional quick demo: hold Alt while loading the page to preload a diagonal VPI18
if (window.location.hash === '#demo-vpi18') {
  const demo = buildDemoVPI18(38);
  clearAllVoxels();
  applyVPI18(demo, updateVoxel);
}

