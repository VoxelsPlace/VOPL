import { HEIGHT, WIDTH, DEPTH } from './palette.js';

export const voxelGrid = Array.from({ length: HEIGHT }, () => Array.from({ length: WIDTH }, () => Array(DEPTH).fill(0)));
export const keyOf = (x, y, z) => `${x},${y},${z}`;
export const voxelMeshes = new Map();
export const pickables = [];
export let selectedColor = 1;
export let currentTool = 'paint';

export function setSelectedColor(i) { selectedColor = i; }
export function setTool(tool) { currentTool = tool; }

