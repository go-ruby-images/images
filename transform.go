// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"fmt"
	"image"
	"math"

	goimages "github.com/go-images/images"
)

// Resize returns the image scaled to width x height using bilinear
// interpolation. Both dimensions must be positive.
func (im *Image) Resize(width, height int) (*Image, error) {
	out, err := goimages.Resize(im.rgba, width, height, goimages.Bilinear)
	if err != nil {
		return nil, fmt.Errorf("images: resize: %w", err)
	}
	return im.wrap(out), nil
}

// ResizeNearest returns the image scaled to width x height using
// nearest-neighbour sampling, which preserves hard pixel edges.
func (im *Image) ResizeNearest(width, height int) (*Image, error) {
	out, err := goimages.Resize(im.rgba, width, height, goimages.NearestNeighbor)
	if err != nil {
		return nil, fmt.Errorf("images: resize: %w", err)
	}
	return im.wrap(out), nil
}

// Thumbnail scales the image to fit inside a maxWidth x maxHeight box while
// preserving the aspect ratio, the equivalent of MiniMagick's `resize
// "WxH"`. The result is never enlarged beyond the original dimensions and each
// side is at least one pixel. Both bounds must be positive.
func (im *Image) Thumbnail(maxWidth, maxHeight int) (*Image, error) {
	if maxWidth <= 0 || maxHeight <= 0 {
		return nil, fmt.Errorf("images: thumbnail: bounds must be positive, got %dx%d", maxWidth, maxHeight)
	}
	w, h := im.Width(), im.Height()
	scale := math.Min(float64(maxWidth)/float64(w), float64(maxHeight)/float64(h))
	if scale > 1 {
		scale = 1
	}
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return im.Resize(nw, nh)
}

// Crop returns the rectangular region of size width x height anchored at
// (x, y). The region must lie fully within the image.
func (im *Image) Crop(x, y, width, height int) (*Image, error) {
	r := image.Rect(x, y, x+width, y+height)
	out, err := goimages.Crop(im.rgba, r)
	if err != nil {
		return nil, fmt.Errorf("images: crop: %w", err)
	}
	return im.wrap(out), nil
}

// Rotate returns the image rotated clockwise by degrees, which must be a
// multiple of 90 (0, 90, 180 or 270 after normalisation).
func (im *Image) Rotate(degrees int) (*Image, error) {
	d := ((degrees % 360) + 360) % 360
	if d%90 != 0 {
		return nil, fmt.Errorf("images: rotate: degrees must be a multiple of 90, got %d", degrees)
	}
	switch d {
	case 0:
		return im.Clone(), nil
	case 90:
		return im.wrap(goimages.Rotate270(im.rgba)), nil
	case 180:
		return im.wrap(goimages.Rotate180(im.rgba)), nil
	default: // 270
		return im.wrap(goimages.Rotate90(im.rgba)), nil
	}
}

// Flip returns the image mirrored top-to-bottom (ImageMagick -flip).
func (im *Image) Flip() *Image {
	return im.wrap(goimages.FlipVertical(im.rgba))
}

// Flop returns the image mirrored left-to-right (ImageMagick -flop).
func (im *Image) Flop() *Image {
	return im.wrap(goimages.FlipHorizontal(im.rgba))
}

// Grayscale returns the image converted to grayscale (Rec. 601 luminance
// replicated across the R, G and B channels).
func (im *Image) Grayscale() *Image {
	return im.wrap(goimages.Grayscale(im.rgba))
}
