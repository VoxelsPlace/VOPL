// Helper to initialize the Go WASM module in browsers and return callable functions
// Usage:
//   <script src="wasm_exec.js"></script>
//   import { initVoplWasm } from './browser.js'
//   const api = await initVoplWasm('vopl.wasm')
//   const voplBytes = api.rle2vopl("256,1,34,7, ...")
//
export async function initVoplWasm(wasmURL = 'vopl.wasm') {
  if (typeof Go === 'undefined') {
    throw new Error('wasm_exec.js not loaded. Copy it from $(go env GOROOT)/misc/wasm/wasm_exec.js');
  }
  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject);
  go.run(result.instance);
  // functions are attached on globalThis by the wasm main
  const { rle2vopl, vopl2glb, packVopls, unpackVoplpack } = globalThis;
  if (!rle2vopl || !vopl2glb || !packVopls || !unpackVoplpack) {
    throw new Error('WASM functions not initialized');
  }
  return { rle2vopl, vopl2glb, packVopls, unpackVoplpack };
}
