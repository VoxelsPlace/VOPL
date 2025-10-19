package vopl

import (
	"fmt"
	"strconv"
)

func ParseHexColor(hex string) ([4]float32, error) {
	if len(hex) == 0 || hex[0] != '#' {
		return [4]float32{}, fmt.Errorf("hex inválido: %s", hex)
	}
	h := hex[1:]
	var r, g, b, a uint64
	var err error
	switch len(h) {
	case 6:
		r, err = strconv.ParseUint(h[0:2], 16, 8)
		g, err = strconv.ParseUint(h[2:4], 16, 8)
		b, err = strconv.ParseUint(h[4:6], 16, 8)
		a = 255
	case 8:
		r, err = strconv.ParseUint(h[0:2], 16, 8)
		g, err = strconv.ParseUint(h[2:4], 16, 8)
		b, err = strconv.ParseUint(h[4:6], 16, 8)
		a, err = strconv.ParseUint(h[6:8], 16, 8)
	default:
		return [4]float32{}, fmt.Errorf("hex length inválido: %s", hex)
	}
	if err != nil {
		return [4]float32{}, err
	}
	return [4]float32{float32(r) / 255, float32(g) / 255, float32(b) / 255, float32(a) / 255}, nil
}
