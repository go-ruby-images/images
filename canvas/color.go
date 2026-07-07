// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

// Package canvas is a pure-Go, chunky_png-style RGBA canvas. Colours are packed
// 0xRRGGBBAA integers (as in ChunkyPNG::Color), pixels are addressable
// individually, primitives (points, lines, rectangles, circles) can be drawn,
// canvases can be alpha-composited, and the whole thing round-trips through PNG.
package canvas

import (
	"fmt"
	"image/color"
)

// Color is a pixel packed as 0xRRGGBBAA, matching ChunkyPNG::Color's integer
// representation so a Ruby binding can pass the same values through unchanged.
type Color uint32

// Common colours, mirroring the chunky_png named constants.
const (
	Transparent Color = 0x00000000
	Black       Color = 0x000000ff
	White       Color = 0xffffffff
	Red         Color = 0xff0000ff
	Green       Color = 0x00ff00ff
	Blue        Color = 0x0000ffff
)

// RGBA builds a colour from 8-bit components.
func RGBA(r, g, b, a uint8) Color {
	return Color(uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a))
}

// RGB builds an opaque colour from 8-bit components.
func RGB(r, g, b uint8) Color { return RGBA(r, g, b, 0xff) }

// R returns the red component.
func (c Color) R() uint8 { return uint8(c >> 24) }

// G returns the green component.
func (c Color) G() uint8 { return uint8(c >> 16) }

// B returns the blue component.
func (c Color) B() uint8 { return uint8(c >> 8) }

// A returns the alpha component.
func (c Color) A() uint8 { return uint8(c) }

// RGBA implements color.Color, returning 16-bit alpha-premultiplied values as
// the image/color model requires.
func (c Color) RGBA() (r, g, b, a uint32) {
	nc := color.NRGBA{R: c.R(), G: c.G(), B: c.B(), A: c.A()}
	return nc.RGBA()
}

// Hex returns the colour as an "#rrggbbaa" string.
func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R(), c.G(), c.B(), c.A())
}

// FromHex parses a "#rgb"-family hex string into a Color. It accepts an
// optional leading '#', and 3 (rgb), 4 (rgba), 6 (rrggbb) or 8 (rrggbbaa)
// hex digits. Missing alpha defaults to fully opaque.
func FromHex(s string) (Color, error) {
	h := s
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	var r, g, b, a uint8 = 0, 0, 0, 0xff
	switch len(h) {
	case 3:
		hi, err := parseNibbles(h)
		if err != nil {
			return 0, err
		}
		r, g, b = hi[0]*0x11, hi[1]*0x11, hi[2]*0x11
	case 4:
		hi, err := parseNibbles(h)
		if err != nil {
			return 0, err
		}
		r, g, b, a = hi[0]*0x11, hi[1]*0x11, hi[2]*0x11, hi[3]*0x11
	case 6:
		by, err := parseBytes(h)
		if err != nil {
			return 0, err
		}
		r, g, b = by[0], by[1], by[2]
	case 8:
		by, err := parseBytes(h)
		if err != nil {
			return 0, err
		}
		r, g, b, a = by[0], by[1], by[2], by[3]
	default:
		return 0, fmt.Errorf("canvas: invalid hex colour %q", s)
	}
	return RGBA(r, g, b, a), nil
}

// parseNibbles decodes each character of h as one hex digit.
func parseNibbles(h string) ([]uint8, error) {
	out := make([]uint8, len(h))
	for i := 0; i < len(h); i++ {
		v, err := hexVal(h[i])
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// parseBytes decodes h as pairs of hex digits into bytes.
func parseBytes(h string) ([]uint8, error) {
	out := make([]uint8, len(h)/2)
	for i := 0; i < len(out); i++ {
		hi, err := hexVal(h[2*i])
		if err != nil {
			return nil, err
		}
		lo, err := hexVal(h[2*i+1])
		if err != nil {
			return nil, err
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

// hexVal returns the numeric value of a single hex digit.
func hexVal(b byte) (uint8, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	default:
		return 0, fmt.Errorf("canvas: invalid hex digit %q", string(b))
	}
}

// Compose alpha-blends fg over bg using straight-alpha "source-over"
// compositing and returns the result, as ChunkyPNG::Color.compose does.
func Compose(fg, bg Color) Color {
	fa := uint32(fg.A())
	if fa == 0xff {
		return fg
	}
	if fa == 0 {
		return bg
	}
	ba := uint32(bg.A())
	inv := 0xff - fa
	// fa is in [1, 254] here (the fully-opaque and fully-transparent cases
	// returned above), so outA is always at least 1.
	outA := fa + ba*inv/0xff
	blend := func(fc, bc uint8) uint8 {
		f := uint32(fc) * fa
		b := uint32(bc) * ba * inv / 0xff
		return uint8((f + b) / outA)
	}
	return RGBA(blend(fg.R(), bg.R()), blend(fg.G(), bg.G()), blend(fg.B(), bg.B()), uint8(outA))
}
