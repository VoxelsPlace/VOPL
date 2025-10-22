export interface VoplWasmAPI {
  vpi2vopl(vpi: Uint8Array): Uint8Array;
  vopl2glb(vopl: Uint8Array): Uint8Array;
  vopl2vpi(vopl: Uint8Array): Uint8Array;
  vpiEncodeEntries(indices: Uint16Array, colors: Uint8Array): Uint8Array;
  vpiDecodeEntries(vpi: Uint8Array): { indices: Uint16Array; colors: Uint8Array };
  packVopls(files: Record<string, Uint8Array>): Uint8Array;
  unpackVoplpack(pack: Uint8Array): Record<string, Uint8Array>;
}
export function initVoplWasm(wasmURL?: string): Promise<VoplWasmAPI>;
