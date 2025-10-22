# VPI18 Specification

**Voxel Packed Index 18-bit Format**
*Version — October 2025*

---

## 1. Introduction

**VPI18** (*Voxel Packed Index 18-bit*) is a compact binary format for representing sparse voxel updates to chunks of fixed size **16×16×16 (4096)**.
It provides **incremental updates**, **efficient storage**, and **constant-time decoding**.

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

| Field     | Bits        | Range  | Description                                |
| --------- | ----------- | ------ | ------------------------------------------ |
| `index`   | 12          | 0–4095 | Linear voxel index inside the chunk        |
| `color`   | 6           | 0–63   | Color index (0 = delete/clear voxel)       |
| **Total** | **18 bits** | —      | Continuous stream of voxel update entries  |

---

## 4. Encoding rules

1. **Color 0 is deletion** — an entry with `color == 0` clears the voxel at `index`.
2. **Continuous packing** — all entries are packed bit-by-bit with no padding.
3. **Entry order is arbitrary** — receivers apply updates in stream order; the last update to the same index wins.
4. **Optional byte alignment** — implementations may align to 3-byte multiples for hardware efficiency.

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

## 7. Incremental updates

Transmit only the changed `(index, color)` pairs using the 18-bit packing scheme. Apply updates directly:

* `color > 0` — set voxel at `index` to `color`.
* `color == 0` — delete/clear voxel at `index`.

---

## 8. Comparison with RLE

| Format    | High Density | Sparse Data  | Diff Support | Parse Simplicity |
| --------- | ------------ | ------------ | ------------ | ---------------- |
| **RLE**   | 🟢 Excellent | 🔴 Poor      | 🔴 None      | 🟢 Simple        |
| **VPI18** | 🟢 Good      | 🟢 Excellent | 🟢 Yes       | 🟠 Binary only   |

---

## 9. Pseudocode

### Encoder (building updates)

```python
for (index, color) in updates:  # updates may include zeros for deletions
    bits = (index << 6) | (color & 0x3F)
    stream.write_bits(bits, 18)
```

### Decoder (applying updates)

```python
while not end_of_stream:
    bits = stream.read_bits(18)
    index = bits >> 6
    color = bits & 0x3F
    if color == 0:
        voxelGrid[index] = 0
    else:
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
