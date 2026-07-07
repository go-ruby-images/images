// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"fmt"

	goimages "github.com/go-images/images"
)

// The methods in this file expose the scikit-image-style operations of the
// sibling github.com/go-images/images engine on the Image type, so a Ruby
// binding can offer them as instance methods. They are thin, deterministic
// adapters: the heavy lifting (separable convolution, morphology, edge
// operators) lives in the reused engine rather than being reimplemented here.

// Invert returns the photographic negative of the image.
func (im *Image) Invert() *Image { return im.wrap(goimages.Invert(im.rgba)) }

// Brightness returns the image with delta added to every channel (clamped).
func (im *Image) Brightness(delta float64) *Image {
	return im.wrap(goimages.AdjustBrightness(im.rgba, delta))
}

// Contrast returns the image with contrast scaled by factor about mid-grey.
func (im *Image) Contrast(factor float64) *Image {
	return im.wrap(goimages.AdjustContrast(im.rgba, factor))
}

// Blur returns a box blur of the given radius (radius must be positive).
func (im *Image) Blur(radius int) (*Image, error) {
	out, err := goimages.BoxBlur(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: blur: %w", err)
	}
	return im.wrap(out), nil
}

// GaussianBlur returns a Gaussian blur with the given sigma (sigma must be
// positive).
func (im *Image) GaussianBlur(sigma float64) (*Image, error) {
	out, err := goimages.GaussianBlur(im.rgba, sigma)
	if err != nil {
		return nil, fmt.Errorf("images: gaussian blur: %w", err)
	}
	return im.wrap(out), nil
}

// Sharpen returns the image sharpened with a fixed 3x3 kernel.
func (im *Image) Sharpen() *Image { return im.wrap(goimages.Sharpen(im.rgba)) }

// UnsharpMask returns an unsharp-masked image (radius must be positive).
func (im *Image) UnsharpMask(radius, amount float64) (*Image, error) {
	out, err := goimages.UnsharpMask(im.rgba, radius, amount)
	if err != nil {
		return nil, fmt.Errorf("images: unsharp mask: %w", err)
	}
	return im.wrap(out), nil
}

// Median returns a median-filtered image over a square window of the given
// radius (radius must be positive).
func (im *Image) Median(radius int) (*Image, error) {
	out, err := goimages.Median(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: median: %w", err)
	}
	return im.wrap(out), nil
}

// Convolve returns the image convolved with an arbitrary odd-sized kernel of
// width x height weights (row-major). It errors on non-positive, even, or
// mismatched dimensions.
func (im *Image) Convolve(width, height int, weights []float64) (*Image, error) {
	out, err := goimages.Convolve(im.rgba, goimages.Kernel{Width: width, Height: height, Weights: weights})
	if err != nil {
		return nil, fmt.Errorf("images: convolve: %w", err)
	}
	return im.wrap(out), nil
}

// Sobel returns the Sobel gradient-magnitude edge map.
func (im *Image) Sobel() *Image { return im.wrap(goimages.Sobel(im.rgba)) }

// Prewitt returns the Prewitt gradient-magnitude edge map.
func (im *Image) Prewitt() *Image { return im.wrap(goimages.Prewitt(im.rgba)) }

// Scharr returns the Scharr gradient-magnitude edge map.
func (im *Image) Scharr() *Image { return im.wrap(goimages.Scharr(im.rgba)) }

// Laplacian returns the Laplacian second-derivative edge map.
func (im *Image) Laplacian() *Image { return im.wrap(goimages.Laplacian(im.rgba)) }

// Canny returns the Canny edge map for the given Gaussian sigma and the low and
// high hysteresis thresholds. It errors on a non-positive sigma.
func (im *Image) Canny(sigma, low, high float64) (*Image, error) {
	out, err := goimages.Canny(im.rgba, sigma, low, high)
	if err != nil {
		return nil, fmt.Errorf("images: canny: %w", err)
	}
	return im.wrap(out), nil
}

// Threshold returns a binary image: pixels whose luminance is at least t become
// white, the rest black.
func (im *Image) Threshold(t uint8) *Image {
	return im.wrap(goimages.Threshold(im.rgba, t))
}

// OtsuThreshold returns the automatically chosen Otsu threshold value.
func (im *Image) OtsuThreshold() uint8 { return goimages.OtsuThreshold(im.rgba) }

// Otsu returns the image binarised at its automatically chosen Otsu threshold.
func (im *Image) Otsu() *Image { return im.wrap(goimages.Otsu(im.rgba)) }

// Erode returns the grayscale morphological erosion over a square structuring
// element of the given radius (radius must be positive).
func (im *Image) Erode(radius int) (*Image, error) {
	out, err := goimages.Erode(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: erode: %w", err)
	}
	return im.wrap(out), nil
}

// Dilate returns the grayscale morphological dilation (radius must be positive).
func (im *Image) Dilate(radius int) (*Image, error) {
	out, err := goimages.Dilate(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: dilate: %w", err)
	}
	return im.wrap(out), nil
}

// Open returns the morphological opening (erosion then dilation).
func (im *Image) Open(radius int) (*Image, error) {
	out, err := goimages.Open(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: open: %w", err)
	}
	return im.wrap(out), nil
}

// Close returns the morphological closing (dilation then erosion).
func (im *Image) Close(radius int) (*Image, error) {
	out, err := goimages.Close(im.rgba, radius)
	if err != nil {
		return nil, fmt.Errorf("images: close: %w", err)
	}
	return im.wrap(out), nil
}
