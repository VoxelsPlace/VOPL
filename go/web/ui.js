import { palette, usableColors } from './palette.js';
import { setSelectedColor, setTool, selectedColor } from './state.js';
import { previewMaterial } from './scene.js';
import { getExamples } from './examples_loader.js';
import { genSquare } from './patterns_export.js';
import { clearAllVoxels } from './patterns_export.js';
import { applyLayers } from './layer_reader.js';
import { getExampleById } from './examples_loader.js';

export function populateColorPalette() {
  const paletteContainer = document.getElementById('colorPalette');
  paletteContainer.innerHTML = '';

  const eraseItem = document.createElement('div');
  eraseItem.classList.add('swatch-item');
  const eraseSwatch = document.createElement('div');
  eraseSwatch.classList.add('color-swatch');
  eraseSwatch.dataset.color = 0;
  eraseSwatch.title = 'Transparent / Erase';
  eraseSwatch.addEventListener('click', () => { setTool('erase'); syncToolButtons(); });
  const eraseLabel = document.createElement('div');
  eraseLabel.classList.add('swatch-label');
  eraseLabel.textContent = '0';
  eraseItem.appendChild(eraseSwatch);
  eraseItem.appendChild(eraseLabel);
  paletteContainer.appendChild(eraseItem);

  for (const i of usableColors) {
    const item = document.createElement('div');
    item.classList.add('swatch-item');
    const swatch = document.createElement('div');
    swatch.classList.add('color-swatch');
    swatch.dataset.color = i;
    swatch.style.backgroundColor = `#${palette[i].toString(16).padStart(6, '0')}`;
    if (i === selectedColor) swatch.classList.add('selected');
    swatch.addEventListener('click', () => {
      setSelectedColor(i);
      setTool('paint');
      const newColorHex = palette[i];
      if (typeof newColorHex === 'number') previewMaterial.color.setHex(newColorHex);
      syncToolButtons();
      syncSelectedSwatch();
    });
    const label = document.createElement('div');
    label.classList.add('swatch-label');
    label.textContent = String(i);
    item.appendChild(swatch);
    item.appendChild(label);
    paletteContainer.appendChild(item);
  }
}

export function syncToolButtons() {
  const paintBtn = document.getElementById('paintTool');
  const eraseBtn = document.getElementById('eraseTool');
  const isPaint = document.body.dataset.tool === 'paint';
  paintBtn.classList.toggle('active', isPaint);
  eraseBtn.classList.toggle('active', !isPaint);
}

export function syncSelectedSwatch() {
  const paletteContainer = document.getElementById('colorPalette');
  const currentSelected = paletteContainer.querySelector('.selected');
  if (currentSelected) currentSelected.classList.remove('selected');
  const swatch = paletteContainer.querySelector(`[data-color="${selectedColor}"]`);
  if (swatch) swatch.classList.add('selected');
}

export function populateExampleButtons(onSelect) {
  const container = document.createElement('div');
  container.classList.add('generator-grid');
  const list = getExamples();
  for (const ex of list) {
    const btn = document.createElement('button');
    btn.textContent = ex.name;
    btn.addEventListener('click', () => onSelect(ex.id));
    container.appendChild(btn);
  }
  return container;
}

