// Helper to initialize the Go WASM module in browsers and return callable functions
// Usage:
//   <script src="wasm_exec.js"></script>
//   import { initVoplWasm } from './browser.js'
//   const api = await initVoplWasm('vopl.wasm')
//   const voplBytes = api.vpi2vopl(vpiBytes)
//
export async function initVoplWasm(wasmURL = 'vopl.wasm') {
  if (typeof Go === 'undefined') {
    throw new Error('wasm_exec.js not loaded. Copy it from $(go env GOROOT)/misc/wasm/wasm_exec.js');
  }
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject);
  go.run(result.instance);
  // functions are attached on globalThis by the wasm main
  const { vpi2vopl, vopl2glb, vopl2vpi, vpiEncodeEntries, vpiDecodeEntries, packVopls, unpackVoplpack } = globalThis;
  if (!vpi2vopl || !vopl2glb || !vopl2vpi || !vpiEncodeEntries || !vpiDecodeEntries || !packVopls || !unpackVoplpack) {
    throw new Error('WASM functions not initialized');
  }
  return { vpi2vopl, vopl2glb, vopl2vpi, vpiEncodeEntries, vpiDecodeEntries, packVopls, unpackVoplpack };
}
