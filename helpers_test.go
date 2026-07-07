// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"bytes"
	"errors"
	"image"
	"image/color"
)

// errWrite is returned by the failing writers used to exercise encoder error
// branches.
var errWrite = errors.New("write failed")

// failWriter fails on every Write, letting tests drive encoder error paths.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errWrite }
func (failWriter) Close() error              { return nil }

// failCloser accepts all writes but fails on Close, exercising the close-error
// branch of the write helpers.
type failCloser struct{ buf bytes.Buffer }

func (f *failCloser) Write(p []byte) (int, error) { return f.buf.Write(p) }
func (f *failCloser) Close() error                { return errWrite }

// gradientRGBA returns a deterministic w x h RGBA gradient anchored at the
// origin.
func gradientRGBA(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 7), uint8(y * 5), uint8((x + y) * 3), 0xff})
		}
	}
	return img
}

// gradientImage returns a gradient as a plain *Image handle.
func gradientImage(w, h int) *Image {
	return &Image{rgba: gradientRGBA(w, h), format: "png"}
}

// encodedPNG returns the PNG-encoded bytes of a small gradient.
func encodedPNG() []byte {
	b, err := gradientImage(4, 4).Convert("png")
	if err != nil {
		panic(err)
	}
	blob, err := b.ToBlob()
	if err != nil {
		panic(err)
	}
	return blob
}
