import { renderer, camera, animate, onResize } from './scene.js';
import { setTool } from './state.js';
import { populateColorPalette, populateExampleButtons } from './ui.js';
import { onPointerMove, onPointerDown } from './input.js';
import { clearAllVoxels, randomNoise, fillChunk, genRainbow, genStripes, genSphere, generateCommand } from './patterns_export.js';
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

document.getElementById('exportRLE').addEventListener('click', generateCommand);
document.getElementById('copyCmd').addEventListener('click', async () => {
  const cmd = generateCommand();
  try {
    await navigator.clipboard.writeText(cmd);
    const btn = document.getElementById('copyCmd');
    const old = btn.innerHTML; btn.innerHTML = 'Copied!';
    setTimeout(() => btn.innerHTML = old, 1500);
  } catch (e) {
    alert('Copy failed. The command is available in the box below.');
  }
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

