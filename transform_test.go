// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import "testing"

func TestResize(t *testing.T) {
	im := gradientImage(8, 6)
	out, err := im.Resize(4, 3)
	if err != nil || out.Width() != 4 || out.Height() != 3 {
		t.Fatalf("Resize = %v, %v", out, err)
	}
	if _, err := im.Resize(0, 3); err == nil {
		t.Fatal("Resize to zero width should fail")
	}
	near, err := im.ResizeNearest(2, 2)
	if err != nil || near.Width() != 2 {
		t.Fatalf("ResizeNearest = %v, %v", near, err)
	}
	if _, err := im.ResizeNearest(2, -1); err == nil {
		t.Fatal("ResizeNearest with bad height should fail")
	}
}

func TestThumbnail(t *testing.T) {
	im := gradientImage(10, 4)
	// Wider than tall: width is the limiting dimension.
	out, err := im.Thumbnail(5, 5)
	if err != nil {
		t.Fatal(err)
	}
	if out.Width() != 5 || out.Height() != 2 {
		t.Fatalf("thumbnail = %dx%d, want 5x2", out.Width(), out.Height())
	}
	// Box larger than the image: never enlarged.
	same, err := im.Thumbnail(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if same.Width() != 10 || same.Height() != 4 {
		t.Fatalf("thumbnail enlarged to %dx%d", same.Width(), same.Height())
	}
	// Extreme aspect ratio: each side stays at least one pixel.
	thin := gradientImage(100, 1)
	tiny, err := thin.Thumbnail(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if tiny.Width() < 1 || tiny.Height() < 1 {
		t.Fatalf("thumbnail collapsed to %dx%d", tiny.Width(), tiny.Height())
	}
	// Tall-and-narrow image: the width would round below one pixel.
	tall := gradientImage(1, 100)
	slim, err := tall.Thumbnail(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if slim.Width() < 1 || slim.Height() < 1 {
		t.Fatalf("thumbnail collapsed to %dx%d", slim.Width(), slim.Height())
	}
	if _, err := im.Thumbnail(0, 5); err == nil {
		t.Fatal("Thumbnail with non-positive bound should fail")
	}
}

func TestCrop(t *testing.T) {
	im := gradientImage(8, 8)
	out, err := im.Crop(2, 2, 4, 3)
	if err != nil || out.Width() != 4 || out.Height() != 3 {
		t.Fatalf("Crop = %v, %v", out, err)
	}
	if _, err := im.Crop(6, 6, 5, 5); err == nil {
		t.Fatal("Crop out of bounds should fail")
	}
}

func TestRotate(t *testing.T) {
	im := gradientImage(4, 2)
	for _, d := range []int{0, 90, 180, 270, 360, -90} {
		if _, err := im.Rotate(d); err != nil {
			t.Fatalf("Rotate(%d): %v", d, err)
		}
	}
	// 90 and 270 swap the axes.
	r90, _ := im.Rotate(90)
	if r90.Width() != 2 || r90.Height() != 4 {
		t.Fatalf("Rotate(90) dims = %dx%d", r90.Width(), r90.Height())
	}
	if _, err := im.Rotate(45); err == nil {
		t.Fatal("Rotate(45) should fail")
	}
}

func TestFlipFlopGrayscale(t *testing.T) {
	im := gradientImage(3, 3)
	if im.Flip().Height() != 3 {
		t.Fatal("Flip changed height")
	}
	if im.Flop().Width() != 3 {
		t.Fatal("Flop changed width")
	}
	g := im.Grayscale()
	px := g.rgba.RGBAAt(1, 1)
	if px.R != px.G || px.G != px.B {
		t.Fatalf("grayscale pixel not neutral: %v", px)
	}
}
