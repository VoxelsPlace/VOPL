export interface VoplHeader {
  ver: number;
  bpp: number;
  w: number;
  h: number;
  d: number;
  pal: number;
  payloadLength: number;
}

export interface VoplDecoded {
  header: VoplHeader;
  grid: Uint8Array; // flattened in (y, x, z) linear order, length = 16*16*16
}

export interface VoplWasmAPI {
  vopl2glb(vopl: Uint8Array): Uint8Array;
  packVopls(files: Record<string, Uint8Array>): Uint8Array;
  unpackVoplpack(pack: Uint8Array): Record<string, Uint8Array>;
  decodeVopl(vopl: Uint8Array): VoplDecoded;
}
export function initVoplWasm(wasmURL?: string): Promise<VoplWasmAPI>;
