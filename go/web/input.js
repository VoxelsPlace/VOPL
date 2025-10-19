import { HEIGHT, WIDTH, DEPTH, palette } from './palette.js';
import { voxelGrid, keyOf, voxelMeshes, pickables, currentTool, selectedColor } from './state.js';
import { scene, previewMaterial, previewEraseMaterial, previewMesh } from './scene.js';

const raycaster = new THREE.Raycaster();
const mouse = new THREE.Vector2();
const materialCache = new Map();

const getMaterial = (colorIdx) => {
  const hex = palette[colorIdx];
  if (typeof hex !== 'number') return null;
  if (!materialCache.has(hex)) {
    materialCache.set(hex, new THREE.MeshStandardMaterial({ color: hex, roughness: 0.7, metalness: 0.1 }));
  }
  return materialCache.get(hex);
};

export function updateVoxel(x, y, z, color) {
  if (x < 0 || x >= WIDTH || y < 0 || y >= HEIGHT || z < 0 || z >= DEPTH) return;
  const k = keyOf(x, y, z);
  if (voxelMeshes.has(k)) {
    const old = voxelMeshes.get(k);
    scene.remove(old);
    old.geometry.dispose();
    const idx = pickables.indexOf(old);
    if (idx !== -1) pickables.splice(idx, 1);
    voxelMeshes.delete(k);
  }
  voxelGrid[y][x][z] = color;
  if (color !== 0) {
    const mat = getMaterial(color);
    if (!mat) return;
    const cube = new THREE.Mesh(new THREE.BoxGeometry(1, 1, 1), mat);
    cube.position.set(x, y, z);
    cube.userData = { x, y, z };
    scene.add(cube);
    voxelMeshes.set(k, cube);
    pickables.push(cube);
  }
}

function getVoxelPlacement(intersect) {
  if (!intersect) return null;
  const pos = new THREE.Vector3().copy(intersect.point);
  const normal = intersect.face?.normal || new THREE.Vector3(0, 1, 0);
  if (intersect.object.userData.isPlane) {
    pos.addScaledVector(normal, -0.01);
    return { x: Math.floor(pos.x + 0.5), y: 0, z: Math.floor(pos.z + 0.5) };
  }
  const { x, y, z } = intersect.object.userData;
  return { x: x + Math.round(normal.x), y: y + Math.round(normal.y), z: z + Math.round(normal.z) };
}

export function onPointerMove(event, renderer, camera) {
  const rect = renderer.domElement.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObjects(pickables, false);
  if (intersects.length > 0) {
    const hit = intersects[0];
    const eraseMode = currentTool === 'erase' || (event.buttons & 2);
    if (hit.object.userData.isPlane && eraseMode) { previewMesh.visible = false; return; }
    let placePos;
    if (eraseMode) {
      if (hit.object.userData.isPlane) { previewMesh.visible = false; return; }
      const { x, y, z } = hit.object.userData;
      placePos = { x, y, z };
      previewMesh.material = previewEraseMaterial;
    } else {
      placePos = getVoxelPlacement(hit);
      previewMesh.material = previewMaterial;
      const hex = palette[selectedColor];
      if (typeof hex === 'number') previewMaterial.color.setHex(hex);
    }
    if (placePos && placePos.x >= 0 && placePos.x < WIDTH && placePos.y >= 0 && placePos.y < HEIGHT && placePos.z >= 0 && placePos.z < DEPTH) {
      previewMesh.position.set(placePos.x, placePos.y, placePos.z);
      previewMesh.visible = true;
    } else {
      previewMesh.visible = false;
    }
  } else {
    previewMesh.visible = false;
  }
}

export function onPointerDown(event, renderer, camera) {
  const rect = renderer.domElement.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObjects(pickables, false);
  if (intersects.length === 0) return;
  const hit = intersects[0];
  const eraseMode = currentTool === 'erase' || event.button === 2;
  if (eraseMode) {
    if (!hit.object.userData.isPlane) {
      const { x, y, z } = hit.object.userData;
      updateVoxel(x, y, z, 0);
    }
  } else {
    const placePos = getVoxelPlacement(hit);
    if (placePos) updateVoxel(placePos.x, placePos.y, placePos.z, selectedColor);
  }
}

