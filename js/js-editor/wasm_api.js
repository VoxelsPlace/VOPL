import { initVoplWasm } from './browser.js';

// Ensure wasm_exec.js is loaded by index.html via <script src="wasm_exec.js"></script>
// Initialize once and export the API for consumers in js-editor.
export const wasmApi = await initVoplWasm('vopl.wasm');
