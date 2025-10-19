package vopl

import "io"

type bitWriter struct {
	buf []byte
	acc uint64
	n   uint8
}

func newBitWriter() *bitWriter { return &bitWriter{buf: make([]byte, 0, 256)} }

func (w *bitWriter) writeBits(v uint64, bits uint8) {
	w.acc |= (v & ((1 << bits) - 1)) << w.n
	w.n += bits
	for w.n >= 8 {
		w.buf = append(w.buf, byte(w.acc&0xFF))
		w.acc >>= 8
		w.n -= 8
	}
}

func (w *bitWriter) bytes() []byte {
	if w.n > 0 {
		w.buf = append(w.buf, byte(w.acc&0xFF))
		w.acc = 0
		w.n = 0
	}
	return w.buf
}

type bitReader struct {
	data []byte
	acc  uint64
	n    uint8
	pos  int
}

func newBitReader(b []byte) *bitReader { return &bitReader{data: b} }

func (r *bitReader) readBits(bits uint8) (uint64, error) {
	for r.n < bits {
		if r.pos >= len(r.data) {
			return 0, io.ErrUnexpectedEOF
		}
		r.acc |= uint64(r.data[r.pos]) << r.n
		r.n += 8
		r.pos++
	}
	mask := uint64((1 << bits) - 1)
	v := r.acc & mask
	r.acc >>= bits
	r.n -= bits
	return v, nil
}

func writeUVarint(dst []byte, x uint32) []byte {
    v := x
    for v >= 0x80 {
        dst = append(dst, byte(v)|0x80)
        v >>= 7
    }
    dst = append(dst, byte(v))
    return dst
}

func readUVarint(src []byte, pos *int) (uint32, error) {
    var x uint32
    var s uint32
    i := *pos
    for {
        if i >= len(src) {
            return 0, io.ErrUnexpectedEOF
        }
        b := src[i]
        i++
        if b < 0x80 {
            if s >= 32 {
                return 0, io.ErrUnexpectedEOF
            }
            x |= uint32(b) << s
            break
        }
        x |= uint32(b&0x7F) << s
        s += 7
        if s > 28 {
            return 0, io.ErrUnexpectedEOF
        }
    }
    *pos = i
    return x, nil
}
