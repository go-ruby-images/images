// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import "testing"

// TestProcessNoError covers every filter that cannot fail, checking that each
// preserves the image dimensions.
func TestProcessNoError(t *testing.T) {
	im := gradientImage(9, 9)
	ops := map[string]*Image{
		"invert":     im.Invert(),
		"brightness": im.Brightness(20),
		"contrast":   im.Contrast(1.5),
		"sharpen":    im.Sharpen(),
		"sobel":      im.Sobel(),
		"prewitt":    im.Prewitt(),
		"scharr":     im.Scharr(),
		"laplacian":  im.Laplacian(),
		"threshold":  im.Threshold(128),
		"otsu":       im.Otsu(),
	}
	for name, out := range ops {
		if out.Width() != 9 || out.Height() != 9 {
			t.Fatalf("%s changed dims to %dx%d", name, out.Width(), out.Height())
		}
	}
	if im.OtsuThreshold() > 255 {
		t.Fatal("impossible threshold")
	}
}

// TestProcessOK covers the fallible filters on valid inputs.
func TestProcessOK(t *testing.T) {
	im := gradientImage(9, 9)
	checks := []struct {
		name string
		fn   func() (*Image, error)
	}{
		{"blur", func() (*Image, error) { return im.Blur(1) }},
		{"gaussian", func() (*Image, error) { return im.GaussianBlur(1.2) }},
		{"unsharp", func() (*Image, error) { return im.UnsharpMask(1, 0.5) }},
		{"median", func() (*Image, error) { return im.Median(1) }},
		{"convolve", func() (*Image, error) { return im.Convolve(3, 3, make([]float64, 9)) }},
		{"canny", func() (*Image, error) { return im.Canny(1.0, 10, 30) }},
		{"erode", func() (*Image, error) { return im.Erode(1) }},
		{"dilate", func() (*Image, error) { return im.Dilate(1) }},
		{"open", func() (*Image, error) { return im.Open(1) }},
		{"close", func() (*Image, error) { return im.Close(1) }},
	}
	for _, c := range checks {
		out, err := c.fn()
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if out.Width() != 9 {
			t.Fatalf("%s changed width", c.name)
		}
	}
}

// TestProcessErrors covers the error branch of every fallible filter.
func TestProcessErrors(t *testing.T) {
	im := gradientImage(9, 9)
	checks := []struct {
		name string
		fn   func() (*Image, error)
	}{
		{"blur", func() (*Image, error) { return im.Blur(0) }},
		{"gaussian", func() (*Image, error) { return im.GaussianBlur(0) }},
		{"unsharp", func() (*Image, error) { return im.UnsharpMask(0, 0.5) }},
		{"median", func() (*Image, error) { return im.Median(0) }},
		{"convolve", func() (*Image, error) { return im.Convolve(2, 2, make([]float64, 4)) }},
		{"canny", func() (*Image, error) { return im.Canny(0, 10, 30) }},
		{"erode", func() (*Image, error) { return im.Erode(0) }},
		{"dilate", func() (*Image, error) { return im.Dilate(0) }},
		{"open", func() (*Image, error) { return im.Open(0) }},
		{"close", func() (*Image, error) { return im.Close(0) }},
	}
	for _, c := range checks {
		if _, err := c.fn(); err == nil {
			t.Fatalf("%s should have errored", c.name)
		}
	}
}
