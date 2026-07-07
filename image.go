// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
)

// Image is the processing handle at the heart of the library — the pure-Go
// analogue of a MiniMagick::Image or a Vips::Image. It wraps an *image.RGBA
// together with the format it was decoded from (or last converted to). All
// transforming methods return a new Image and never mutate the receiver, so a
// handle can be shared safely.
type Image struct {
	rgba   *image.RGBA
	format string
}

// toRGBA converts any image.Image to *image.RGBA, returning the input
// unchanged when it is already an origin-anchored RGBA image.
func toRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok && rgba.Rect.Min == (image.Point{}) {
		return rgba
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(x, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

// New returns a fresh Image of the given size filled with fill. The default
// output format is PNG. It is the entry point for building an image from
// scratch, mirroring Vips::Image.black plus a fill.
func New(width, height int, fill color.Color) *Image {
	rgba := image.NewRGBA(image.Rect(0, 0, max0(width), max0(height)))
	r, g, b, a := fill.RGBA()
	c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
	for i := 0; i+3 < len(rgba.Pix); i += 4 {
		rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], rgba.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	return &Image{rgba: rgba, format: "png"}
}

// FromImage wraps an existing image.Image, converting it to RGBA. The reported
// format defaults to PNG. It bridges the standard-library image world into the
// Image API.
func FromImage(src image.Image, format string) *Image {
	f := normalizeFormat(format)
	if f == "" {
		f = "png"
	}
	return &Image{rgba: toRGBA(src), format: f}
}

// Decode reads and decodes an image from r, auto-detecting the format.
func Decode(r io.Reader) (*Image, error) {
	rgba, format, err := decode(r)
	if err != nil {
		return nil, err
	}
	return &Image{rgba: rgba, format: format}, nil
}

// Read decodes an image from an in-memory byte slice, the equivalent of
// MiniMagick::Image.read(blob).
func Read(data []byte) (*Image, error) {
	return Decode(bytes.NewReader(data))
}

// Open reads and decodes the image file at path.
func Open(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("images: open: %w", err)
	}
	defer f.Close()
	img, err := Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// RGBA returns the underlying *image.RGBA. Callers that mutate it affect this
// Image; use Clone first to keep the original intact.
func (im *Image) RGBA() *image.RGBA { return im.rgba }

// Clone returns a deep copy of the image.
func (im *Image) Clone() *Image {
	dup := image.NewRGBA(im.rgba.Bounds())
	copy(dup.Pix, im.rgba.Pix)
	return &Image{rgba: dup, format: im.format}
}

// Width returns the image width in pixels.
func (im *Image) Width() int { return im.rgba.Bounds().Dx() }

// Height returns the image height in pixels.
func (im *Image) Height() int { return im.rgba.Bounds().Dy() }

// Dimensions returns the width and height in pixels.
func (im *Image) Dimensions() (int, int) { return im.Width(), im.Height() }

// Format returns the current canonical format name (for example "png").
func (im *Image) Format() string { return im.format }

// At returns the colour at (x, y). It reports an error when the coordinates
// fall outside the image bounds.
func (im *Image) At(x, y int) (color.RGBA, error) {
	if x < 0 || y < 0 || x >= im.Width() || y >= im.Height() {
		return color.RGBA{}, fmt.Errorf("images: At(%d, %d) out of bounds %dx%d", x, y, im.Width(), im.Height())
	}
	return im.rgba.RGBAAt(x, y), nil
}

// Metadata bundles the read-only descriptive properties of an image.
type Metadata struct {
	Width      int
	Height     int
	Format     string
	ColorModel string
}

// Metadata returns the descriptive properties of the image.
func (im *Image) Metadata() Metadata {
	return Metadata{
		Width:      im.Width(),
		Height:     im.Height(),
		Format:     im.format,
		ColorModel: "rgba",
	}
}

// Convert changes the output format used by Encode, ToBlob and Write (when the
// path carries no extension). It returns an error for formats this package
// cannot encode.
func (im *Image) Convert(format string) (*Image, error) {
	f := normalizeFormat(format)
	for _, e := range Encodable() {
		if e == f {
			out := im.Clone()
			out.format = f
			return out, nil
		}
	}
	return nil, fmt.Errorf("images: convert: cannot encode format %q", format)
}

// Encode writes the image to w in its current format.
func (im *Image) Encode(w io.Writer) error {
	return encode(w, im.rgba, im.format)
}

// ToBlob encodes the image in its current format and returns the bytes, the
// equivalent of MiniMagick::Image#to_blob.
func (im *Image) ToBlob() ([]byte, error) {
	var buf bytes.Buffer
	if err := im.Encode(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Write encodes the image to path. The format is taken from the path extension
// when present, otherwise the image's current format is used.
func (im *Image) Write(path string) error {
	format := im.format
	if f, err := formatFromExt(path); err == nil {
		format = f
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("images: write: %w", err)
	}
	return writeTo(f, im.rgba, format)
}

// writeTo encodes img to wc in the given format and closes wc. It is factored
// out so both the encode failure and the close failure can be exercised.
func writeTo(wc io.WriteCloser, img image.Image, format string) error {
	if err := encode(wc, img, format); err != nil {
		wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("images: write: %w", err)
	}
	return nil
}

// wrap builds a new Image from a processed RGBA result, preserving the format.
func (im *Image) wrap(rgba *image.RGBA) *Image {
	return &Image{rgba: rgba, format: im.format}
}

// max0 clamps n to a minimum of zero.
func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}
