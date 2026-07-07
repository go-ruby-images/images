// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package canvas

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"io"
	"path/filepath"
	"testing"
)

var errBoom = errors.New("boom")

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errBoom }
func (failWriter) Close() error              { return nil }

type failCloser struct{ buf bytes.Buffer }

func (f *failCloser) Write(p []byte) (int, error) { return f.buf.Write(p) }
func (f *failCloser) Close() error                { return errBoom }

func TestNewAndAccess(t *testing.T) {
	c := New(3, 2, Red)
	if c.Width() != 3 || c.Height() != 2 {
		t.Fatalf("dims = %dx%d", c.Width(), c.Height())
	}
	got, err := c.At(2, 1)
	if err != nil || got != Red {
		t.Fatalf("At = %08x, %v", uint32(got), err)
	}
	if err := c.Set(0, 0, Blue); err != nil {
		t.Fatal(err)
	}
	if err := c.Point(1, 1, Green); err != nil {
		t.Fatal(err)
	}
	// Out-of-bounds access on both axes and both directions.
	for _, p := range [][2]int{{-1, 0}, {0, -1}, {3, 0}, {0, 2}} {
		if _, err := c.At(p[0], p[1]); err == nil {
			t.Fatalf("At%v should fail", p)
		}
		if err := c.Set(p[0], p[1], Red); err == nil {
			t.Fatalf("Set%v should fail", p)
		}
	}
	// Negative dimensions clamp to an empty canvas.
	if z := New(-2, -3, White); z.Width() != 0 || z.Height() != 0 {
		t.Fatalf("negative canvas = %dx%d", z.Width(), z.Height())
	}
}

func TestFill(t *testing.T) {
	c := New(2, 2, Black)
	c.Fill(White)
	got, _ := c.At(1, 1)
	if got != White {
		t.Fatalf("Fill left %08x", uint32(got))
	}
}

func TestLine(t *testing.T) {
	c := New(5, 5, Black)
	c.Line(0, 0, 4, 4, White) // diagonal exercises both Bresenham steps
	c.Line(0, 2, 4, 2, White) // horizontal
	c.Line(2, 0, 2, 4, White) // vertical
	c.Line(0, 0, 4, 1, White) // shallow slope
	c.Line(1, 1, 1, 1, White) // single point (immediate return)
	c.Line(4, 0, 0, 4, White) // reversed direction (negative deltas)
	c.Line(3, 3, 8, 8, White) // runs off the canvas and is clipped
	if got, _ := c.At(0, 0); got != White {
		t.Fatal("diagonal endpoint not drawn")
	}
	if got, _ := c.At(2, 2); got != White {
		t.Fatal("center not drawn")
	}
}

func TestRect(t *testing.T) {
	c := New(6, 6, Black)
	c.Rect(1, 1, 4, 4, White, Red) // filled
	center, _ := c.At(2, 2)
	if center != Red {
		t.Fatalf("fill = %08x", uint32(center))
	}
	edge, _ := c.At(1, 1)
	if edge != White {
		t.Fatalf("stroke = %08x", uint32(edge))
	}
	// Swapped corners and a transparent (no-op) fill.
	c.Rect(5, 5, 2, 2, Green, Transparent)
	// Partly off-canvas: exercises the clip path.
	c.Rect(-2, -2, 2, 2, Blue, Blue)
}

func TestCircle(t *testing.T) {
	c := New(9, 9, Black)
	c.Circle(4, 4, 3, White)
	if got, _ := c.At(4, 1); got != White {
		t.Fatalf("circle top = %08x", uint32(got))
	}
	c.Circle(0, 0, 4, White)  // clipped against the corner
	c.Circle(2, 2, -1, White) // negative radius is a no-op
}

func TestCanvasCompose(t *testing.T) {
	base := New(4, 4, Blue)
	top := New(2, 2, RGBA(0xff, 0, 0, 0x80))
	base.Compose(top, 1, 1) // fully inside
	base.Compose(top, 3, 3) // partly off-canvas -> clipped
	got, _ := base.At(1, 1)
	if got.R() == 0 {
		t.Fatalf("compose did not blend: %08x", uint32(got))
	}
}

func TestToRGBAAndFromImage(t *testing.T) {
	c := New(3, 2, RGBA(1, 2, 3, 4))
	img := c.ToRGBA()
	if img.Bounds().Dx() != 3 || img.Bounds().Dy() != 2 {
		t.Fatalf("ToRGBA bounds %v", img.Bounds())
	}
	back := FromImage(img)
	got, _ := back.At(0, 0)
	if got != RGBA(1, 2, 3, 4) {
		t.Fatalf("round trip = %08x", uint32(got))
	}
	// FromImage from a non-RGBA source.
	nr := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	nr.SetNRGBA(0, 0, color.NRGBA{9, 8, 7, 255})
	if FromImage(nr).Width() != 1 {
		t.Fatal("FromImage NRGBA width")
	}
}

func TestPNGRoundTrip(t *testing.T) {
	c := New(3, 3, Green)
	var buf bytes.Buffer
	if err := c.EncodePNG(&buf); err != nil {
		t.Fatal(err)
	}
	dec, err := DecodePNG(bytes.NewReader(buf.Bytes()))
	if err != nil || dec.Width() != 3 {
		t.Fatalf("DecodePNG = %v, %v", dec, err)
	}
	if _, err := DecodePNG(bytes.NewReader([]byte("nope"))); err == nil {
		t.Fatal("DecodePNG of garbage should fail")
	}

	blob, err := c.ToBlob()
	if err != nil || len(blob) == 0 {
		t.Fatalf("ToBlob = %d, %v", len(blob), err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "c.png")
	if err := c.SavePNG(path); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPNG(path); err != nil {
		t.Fatalf("LoadPNG: %v", err)
	}
	if _, err := LoadPNG(filepath.Join(dir, "missing.png")); err == nil {
		t.Fatal("LoadPNG missing should fail")
	}
	if err := c.SavePNG(filepath.Join(dir, "nope", "c.png")); err == nil {
		t.Fatal("SavePNG to missing dir should fail")
	}
}

func TestEncodeErrorPaths(t *testing.T) {
	c := New(2, 2, White)
	// A failing writer trips EncodePNG's error branch.
	if err := c.EncodePNG(failWriter{}); err == nil {
		t.Fatal("EncodePNG to failing writer should error")
	}
	// saveTo's encode-failure and close-failure branches.
	if err := c.saveTo(failWriter{}); err == nil {
		t.Fatal("saveTo encode failure should error")
	}
	if err := c.saveTo(&failCloser{}); err == nil {
		t.Fatal("saveTo close failure should error")
	}
	// ToBlob's in-memory error branch, reachable only via the encoder seam.
	orig := pngEncode
	pngEncode = func(io.Writer, image.Image) error { return errBoom }
	defer func() { pngEncode = orig }()
	if _, err := c.ToBlob(); err == nil {
		t.Fatal("ToBlob should surface the encoder error")
	}
}
