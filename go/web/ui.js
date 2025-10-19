import { palette, usableColors } from './palette.js';
import { setSelectedColor, setTool, selectedColor } from './state.js';
import { previewMaterial } from './scene.js';

export function populateColorPalette() {
  const paletteContainer = document.getElementById('colorPalette');
  paletteContainer.innerHTML = '';

  const eraseSwatch = document.createElement('div');
  eraseSwatch.classList.add('color-swatch');
  eraseSwatch.dataset.color = 0;
  eraseSwatch.title = 'Transparent / Erase';
  eraseSwatch.addEventListener('click', () => { setTool('erase'); syncToolButtons(); });
  paletteContainer.appendChild(eraseSwatch);

  for (const i of usableColors) {
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
    paletteContainer.appendChild(swatch);
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

