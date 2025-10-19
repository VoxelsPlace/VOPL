import { HEIGHT, WIDTH, DEPTH, palette } from './palette.js';
import { pickables } from './state.js';

export const scene = new THREE.Scene();
scene.background = new THREE.Color(getComputedStyle(document.body).getPropertyValue('--bg-color'));

export const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 0.1, 1000);
export const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.body.appendChild(renderer.domElement);

export const controls = new THREE.OrbitControls(camera, renderer.domElement);
controls.target.set(WIDTH / 2 - 0.5, HEIGHT / 2 - 0.5, DEPTH / 2 - 0.5);
camera.position.set(WIDTH * 1.2, HEIGHT * 1.8, DEPTH * 1.6);

const ambientLight = new THREE.AmbientLight(0xffffff, 0.7);
scene.add(ambientLight);
const directionalLight = new THREE.DirectionalLight(0xffffff, 0.5);
directionalLight.position.set(5, 10, 7.5);
scene.add(directionalLight);

const gridHelper = new THREE.GridHelper(Math.max(WIDTH, DEPTH), Math.max(WIDTH, DEPTH));
gridHelper.position.set(WIDTH / 2 - 0.5, -0.5, DEPTH / 2 - 0.5);
scene.add(gridHelper);

export const plane = new THREE.Mesh(
  new THREE.PlaneGeometry(WIDTH, DEPTH),
  new THREE.MeshBasicMaterial({ visible: false })
);
plane.rotation.x = -Math.PI / 2;
plane.position.set(WIDTH / 2 - 0.5, -0.5, DEPTH / 2 - 0.5);
plane.userData.isPlane = true;
scene.add(plane);
pickables.push(plane);

export const previewMaterial = new THREE.MeshBasicMaterial({ color: palette[1], transparent: true, opacity: 0.5 });
export const previewEraseMaterial = new THREE.MeshBasicMaterial({ color: 0xff0000, transparent: true, opacity: 0.4 });
export const previewMesh = new THREE.Mesh(new THREE.BoxGeometry(1.01, 1.01, 1.01), previewMaterial);
previewMesh.visible = false;
scene.add(previewMesh);

export function animate() {
  requestAnimationFrame(animate);
  controls.update();
  renderer.render(scene, camera);
}

export function onResize() {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
}

