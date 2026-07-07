// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package canvas

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

// Canvas is a mutable rectangular grid of [Color] pixels, the pure-Go analogue
// of ChunkyPNG::Canvas. The origin (0, 0) is the top-left corner.
type Canvas struct {
	width  int
	height int
	pixels []Color
}

// New returns a width x height canvas with every pixel set to fill. Negative
// dimensions are clamped to zero.
func New(width, height int, fill Color) *Canvas {
	w, h := clamp0(width), clamp0(height)
	px := make([]Color, w*h)
	for i := range px {
		px[i] = fill
	}
	return &Canvas{width: w, height: h, pixels: px}
}

// clamp0 clamps n to a minimum of zero.
func clamp0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// Width returns the canvas width in pixels.
func (c *Canvas) Width() int { return c.width }

// Height returns the canvas height in pixels.
func (c *Canvas) Height() int { return c.height }

// inBounds reports whether (x, y) lies inside the canvas.
func (c *Canvas) inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < c.width && y < c.height
}

// At returns the colour at (x, y), erroring when out of bounds.
func (c *Canvas) At(x, y int) (Color, error) {
	if !c.inBounds(x, y) {
		return 0, fmt.Errorf("canvas: At(%d, %d) out of bounds %dx%d", x, y, c.width, c.height)
	}
	return c.pixels[y*c.width+x], nil
}

// Set writes col at (x, y), erroring when out of bounds.
func (c *Canvas) Set(x, y int, col Color) error {
	if !c.inBounds(x, y) {
		return fmt.Errorf("canvas: Set(%d, %d) out of bounds %dx%d", x, y, c.width, c.height)
	}
	c.pixels[y*c.width+x] = col
	return nil
}

// Point is an alias for Set, matching ChunkyPNG's drawing vocabulary.
func (c *Canvas) Point(x, y int, col Color) error { return c.Set(x, y, col) }

// setClipped writes col at (x, y) only when it lies inside the canvas.
func (c *Canvas) setClipped(x, y int, col Color) {
	if c.inBounds(x, y) {
		c.pixels[y*c.width+x] = col
	}
}

// Fill sets every pixel to col.
func (c *Canvas) Fill(col Color) {
	for i := range c.pixels {
		c.pixels[i] = col
	}
}

// Line draws a straight line from (x0, y0) to (x1, y1) in col using Bresenham's
// algorithm. Pixels outside the canvas are silently clipped.
func (c *Canvas) Line(x0, y0, x1, y1 int, col Color) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := step(x0, x1)
	sy := step(y0, y1)
	err := dx + dy
	for {
		c.setClipped(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// Rect draws an axis-aligned rectangle with corners (x0, y0) and (x1, y1). The
// border is drawn in stroke; if fill's alpha is non-zero the interior is filled
// with fill. Pixels outside the canvas are clipped.
func (c *Canvas) Rect(x0, y0, x1, y1 int, stroke, fill Color) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	if fill.A() != 0 {
		for y := y0 + 1; y < y1; y++ {
			for x := x0 + 1; x < x1; x++ {
				c.setClipped(x, y, fill)
			}
		}
	}
	c.Line(x0, y0, x1, y0, stroke)
	c.Line(x0, y1, x1, y1, stroke)
	c.Line(x0, y0, x0, y1, stroke)
	c.Line(x1, y0, x1, y1, stroke)
}

// Circle draws a circle of radius r centred at (cx, cy) in col using the
// midpoint algorithm. Pixels outside the canvas are clipped.
func (c *Canvas) Circle(cx, cy, r int, col Color) {
	if r < 0 {
		return
	}
	x, y := r, 0
	err := 1 - r
	for x >= y {
		c.setClipped(cx+x, cy+y, col)
		c.setClipped(cx+y, cy+x, col)
		c.setClipped(cx-y, cy+x, col)
		c.setClipped(cx-x, cy+y, col)
		c.setClipped(cx-x, cy-y, col)
		c.setClipped(cx-y, cy-x, col)
		c.setClipped(cx+y, cy-x, col)
		c.setClipped(cx+x, cy-y, col)
		y++
		if err < 0 {
			err += 2*y + 1
		} else {
			x--
			err += 2*(y-x) + 1
		}
	}
}

// Compose alpha-blends other onto c with its top-left corner at (x, y), using
// source-over compositing. The overlapping region is clipped to c's bounds.
func (c *Canvas) Compose(other *Canvas, x, y int) {
	for oy := 0; oy < other.height; oy++ {
		for ox := 0; ox < other.width; ox++ {
			dx, dy := x+ox, y+oy
			if !c.inBounds(dx, dy) {
				continue
			}
			idx := dy*c.width + dx
			c.pixels[idx] = Compose(other.pixels[oy*other.width+ox], c.pixels[idx])
		}
	}
}

// ToRGBA renders the canvas into a fresh *image.RGBA.
func (c *Canvas) ToRGBA() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	for i, p := range c.pixels {
		o := i * 4
		img.Pix[o] = p.R()
		img.Pix[o+1] = p.G()
		img.Pix[o+2] = p.B()
		img.Pix[o+3] = p.A()
	}
	return img
}

// FromImage builds a canvas from any image.Image.
func FromImage(src image.Image) *Canvas {
	b := src.Bounds()
	cv := New(b.Dx(), b.Dy(), Transparent)
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, a := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			cv.pixels[y*cv.width+x] = RGBA(uint8(r>>8), uint8(g>>8), uint8(bl>>8), uint8(a>>8))
		}
	}
	return cv
}

// DecodePNG reads a PNG image from r into a canvas.
func DecodePNG(r io.Reader) (*Canvas, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("canvas: decode png: %w", err)
	}
	return FromImage(img), nil
}

// LoadPNG reads the PNG file at path into a canvas.
func LoadPNG(path string) (*Canvas, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("canvas: load: %w", err)
	}
	defer f.Close()
	return DecodePNG(f)
}

// pngEncode is the PNG encoder used by EncodePNG. It is a package variable so
// tests can substitute a failing encoder to exercise the in-memory error path
// of ToBlob, which a bytes.Buffer alone can never trigger.
var pngEncode = png.Encode

// EncodePNG writes the canvas to w as a PNG.
func (c *Canvas) EncodePNG(w io.Writer) error {
	if err := pngEncode(w, c.ToRGBA()); err != nil {
		return fmt.Errorf("canvas: encode png: %w", err)
	}
	return nil
}

// ToBlob encodes the canvas as PNG bytes.
func (c *Canvas) ToBlob() ([]byte, error) {
	var buf bytes.Buffer
	if err := c.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SavePNG writes the canvas to path as a PNG.
func (c *Canvas) SavePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("canvas: save: %w", err)
	}
	return c.saveTo(f)
}

// saveTo encodes the canvas to wc and closes it. It is factored out so both the
// encode failure and the close failure can be exercised in tests.
func (c *Canvas) saveTo(wc io.WriteCloser) error {
	if err := c.EncodePNG(wc); err != nil {
		wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("canvas: save: %w", err)
	}
	return nil
}

// abs returns the absolute value of n.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// step returns the sign of (to - from): -1, 0 or +1.
func step(from, to int) int {
	if from < to {
		return 1
	}
	if from > to {
		return -1
	}
	return 0
}
