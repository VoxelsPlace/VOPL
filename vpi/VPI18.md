# VPI18 Specification

**Voxel Packed Index 18-bit Format**
*Version 1.0 — October 2025*

---

## 1. Introduction

**VPI18** (*Voxel Packed Index 18-bit*) is a compact binary format for representing voxel chunks of fixed size **16×16×16 (4096)**.
It replaces traditional RLE by providing **incremental updates**, **efficient storage**, and **constant-time decoding**.

Each active voxel (≠ 0) is encoded in **18 bits**, combining **12 bits for the voxel’s linear index** and **6 bits for the color index**.

---

## 2. Chunk structure

* Each **chunk** contains up to 4096 voxels.
* The **linear index** is derived from the 3D coordinates using fixed traversal order `(x, y, z)` within a 16³ grid:

  ```
  index = x + y*16 + z*256
  ```
* Index range: `0` (bottom-left corner) → `4095` (top-right corner).

---

## 3. Binary layout

| Field     | Bits        | Range  | Description                          |
| --------- | ----------- | ------ | ------------------------------------ |
| `index`   | 12          | 0–4095 | Linear voxel index inside the chunk  |
| `color`   | 6           | 0–63   | Color index (0 = empty/air, ignored) |
| **Total** | **18 bits** | —      | Continuous stream of active voxels   |

---

## 4. Encoding rules

1. **Skip empty voxels** — any voxel with `color == 0` is omitted.
2. **Continuous packing** — all voxels are packed bit-by-bit with no padding.
3. **Fixed traversal order:** iterate `y → z → x`.
4. **Optional byte alignment:** implementations may align to 3-byte multiples for hardware efficiency.

---

## 5. Example

### 5.1 Input RLE

```
17,0,2,7,10,0,2,1,2,0,1,7,12,0,1,1,162,0
```

### 5.2 Decoded voxels

| Index | Color |
| ----- | ----- |
| 17    | 7     |
| 30    | 1     |
| 45    | 7     |
| 58    | 1     |
| 234   | 1     |

### 5.3 Output VPI18 (bytes)

```
[0x04, 0x71, 0xC0, 0x1E, 0x06, 0x80, 0x3A, 0xE1, 0x00]
```

Each 3-byte group (24 bits) contains 18 valid bits; the remaining 6 bits spill into the next voxel’s data.

---

## 6. Optional header (future extension)

| Field         | Size    | Type    | Description               |
| ------------- | ------- | ------- | ------------------------- |
| `magic`       | 4 bytes | char[4] | `"VPI1"`                  |
| `version`     | 1 byte  | uint8   | Format version            |
| `chunkIndex`  | 4 bytes | uint32  | Linear chunk ID           |
| `payloadSize` | 4 bytes | uint32  | Bitstream length in bytes |

> Raw streams (for networking or diff updates) may omit this header entirely.

---

## 7. Incremental updates (VPI18-Diff)

For differential updates:

* Transmit only the changed `(index, color)` pairs.
* Same 18-bit packing scheme.
* The client applies updates directly via `updateVoxel(index, color)` without rebuilding the full chunk.

---

## 8. Comparison with RLE

| Format    | High Density | Sparse Data  | Diff Support | Parse Simplicity |
| --------- | ------------ | ------------ | ------------ | ---------------- |
| **RLE**   | 🟢 Excellent | 🔴 Poor      | 🔴 None      | 🟢 Simple        |
| **VPI18** | 🟢 Good      | 🟢 Excellent | 🟢 Yes       | 🟠 Binary only   |

---

## 9. Pseudocode

### Encoder

```python
for (index, color) in voxels:
    if color == 0:
        continue
    bits = (index << 6) | color
    stream.write_bits(bits, 18)
```

### Decoder

```python
while not end_of_stream:
    bits = stream.read_bits(18)
    index = bits >> 6
    color = bits & 0x3F
    voxelGrid[index] = color
```

---

## 10. Maximum sizes

| Chunk           | Active Voxels | Size                 |
| --------------- | ------------- | -------------------- |
| 16×16×16 (4096) | 4096          | 9,216 bytes (9.0 KB) |
| 10% filled      | ~410          | ~0.9 KB              |
| 1% filled       | ~40           | ~90 bytes            |

---

## 11. Summary

**VPI18** delivers:

* Efficient packing of voxel data.
* Straightforward bit-level encoding.
* Incremental update capability.
* Compatibility with real-time voxel engines using 16³ chunks.
