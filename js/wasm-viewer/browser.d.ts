export interface VoplWasmAPI {
  rle2vopl(rle: string): Uint8Array;
  vopl2glb(vopl: Uint8Array): Uint8Array;
  packVopls(files: Record<string, Uint8Array>): Uint8Array;
  unpackVoplpack(pack: Uint8Array): Record<string, Uint8Array>;
}
export function initVoplWasm(wasmURL?: string): Promise<VoplWasmAPI>;
