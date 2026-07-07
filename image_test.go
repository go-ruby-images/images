// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	im := New(3, 2, color.RGBA{10, 20, 30, 40})
	if im.Width() != 3 || im.Height() != 2 {
		t.Fatalf("dims = %dx%d, want 3x2", im.Width(), im.Height())
	}
	if im.Format() != "png" {
		t.Fatalf("format = %q, want png", im.Format())
	}
	got, err := im.At(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != (color.RGBA{10, 20, 30, 40}) {
		t.Fatalf("pixel = %v", got)
	}
	// Negative dimensions clamp to zero.
	if z := New(-4, -1, color.Black); z.Width() != 0 || z.Height() != 0 {
		t.Fatalf("negative dims not clamped: %dx%d", z.Width(), z.Height())
	}
}

func TestFromImageAndToRGBA(t *testing.T) {
	// Passthrough branch: origin-anchored RGBA is reused.
	src := gradientRGBA(2, 2)
	if FromImage(src, "png").RGBA() != src {
		t.Fatal("origin RGBA should be reused")
	}
	// Default format branch (empty format -> png).
	if FromImage(src, "").Format() != "png" {
		t.Fatal("empty format should default to png")
	}
	// Alias normalisation branch.
	if FromImage(src, "jpg").Format() != "jpeg" {
		t.Fatal("jpg should normalise to jpeg")
	}
	// Conversion branch: a non-RGBA image is copied.
	nrgba := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	nrgba.SetNRGBA(0, 0, color.NRGBA{1, 2, 3, 255})
	if got := FromImage(nrgba, "png"); got.Width() != 2 {
		t.Fatalf("nrgba width = %d", got.Width())
	}
	// Non-origin RGBA also goes through the conversion branch.
	off := gradientRGBA(4, 4).SubImage(image.Rect(1, 1, 3, 3)).(*image.RGBA)
	if got := FromImage(off, "png"); got.Width() != 2 || got.Height() != 2 {
		t.Fatalf("subimage dims = %dx%d", got.Width(), got.Height())
	}
}

func TestDecodeReadOpen(t *testing.T) {
	blob := encodedPNG()
	im, err := Decode(bytes.NewReader(blob))
	if err != nil {
		t.Fatal(err)
	}
	if im.Format() != "png" || im.Width() != 4 {
		t.Fatalf("decoded %s %dx%d", im.Format(), im.Width(), im.Height())
	}
	if _, err := Read(blob); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if _, err := Read([]byte("not an image")); err == nil {
		t.Fatal("Read of garbage should fail")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "g.png")
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := Open(filepath.Join(dir, "missing.png")); err == nil {
		t.Fatal("Open of missing file should fail")
	}
	bad := filepath.Join(dir, "bad.png")
	if err := os.WriteFile(bad, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(bad); err == nil {
		t.Fatal("Open of non-image should fail")
	}
}

func TestAtBounds(t *testing.T) {
	im := gradientImage(2, 2)
	if _, err := im.At(0, 0); err != nil {
		t.Fatal(err)
	}
	for _, p := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 2}} {
		if _, err := im.At(p[0], p[1]); err == nil {
			t.Fatalf("At%v should be out of bounds", p)
		}
	}
}

func TestMetadataAndClone(t *testing.T) {
	im := gradientImage(5, 3)
	m := im.Metadata()
	if m.Width != 5 || m.Height != 3 || m.Format != "png" || m.ColorModel != "rgba" {
		t.Fatalf("metadata = %+v", m)
	}
	w, h := im.Dimensions()
	if w != 5 || h != 3 {
		t.Fatalf("dimensions = %dx%d", w, h)
	}
	clone := im.Clone()
	clone.rgba.Pix[3] = 0 // mutate the alpha of pixel (0,0), originally 0xff
	if im.rgba.Pix[3] != 0xff {
		t.Fatal("clone shares backing store")
	}
}

func TestConvert(t *testing.T) {
	im := gradientImage(2, 2)
	for _, f := range Encodable() {
		out, err := im.Convert(f)
		if err != nil || out.Format() != f {
			t.Fatalf("Convert(%q) = %q, %v", f, out.Format(), err)
		}
	}
	// Alias normalisation reaches the tiff branch.
	if out, err := im.Convert("tif"); err != nil || out.Format() != "tiff" {
		t.Fatalf("Convert(tif) = %v, %v", out, err)
	}
	if _, err := im.Convert("webp"); err == nil {
		t.Fatal("Convert(webp) should fail")
	}
	if _, err := im.Convert("xcf"); err == nil {
		t.Fatal("Convert(xcf) should fail")
	}
}

func TestEncodeToBlobAllFormats(t *testing.T) {
	im := gradientImage(4, 4)
	for _, f := range Encodable() {
		conv, err := im.Convert(f)
		if err != nil {
			t.Fatal(err)
		}
		blob, err := conv.ToBlob()
		if err != nil || len(blob) == 0 {
			t.Fatalf("ToBlob(%q) = %d bytes, %v", f, len(blob), err)
		}
		// Every encodable format is also decodable, closing the round trip.
		if _, err := Read(blob); err != nil {
			t.Fatalf("round trip %q: %v", f, err)
		}
	}
}

func TestEncodeErrorBranches(t *testing.T) {
	im := gradientImage(3, 3)
	for _, f := range Encodable() {
		conv, _ := im.Convert(f)
		if err := conv.Encode(failWriter{}); err == nil {
			t.Fatalf("Encode(%q) to failing writer should error", f)
		}
	}
	// webp encoding and an unknown format are rejected without a writer.
	webp := &Image{rgba: im.rgba, format: "webp"}
	if err := webp.Encode(&bytes.Buffer{}); err == nil {
		t.Fatal("webp encode should fail")
	}
	if _, err := webp.ToBlob(); err == nil {
		t.Fatal("webp ToBlob should fail")
	}
	unknown := &Image{rgba: im.rgba, format: "zzz"}
	if err := unknown.Encode(&bytes.Buffer{}); err == nil {
		t.Fatal("unknown-format encode should fail")
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	im := gradientImage(4, 4)
	// Extension drives the format.
	if err := im.Write(filepath.Join(dir, "out.gif")); err != nil {
		t.Fatalf("Write gif: %v", err)
	}
	// No extension: fall back to the current format.
	if err := im.Write(filepath.Join(dir, "noext")); err != nil {
		t.Fatalf("Write noext: %v", err)
	}
	// Trailing-dot and unsupported extensions still fall back to png.
	if err := im.Write(filepath.Join(dir, "trailing.")); err != nil {
		t.Fatalf("Write trailing dot: %v", err)
	}
	if err := im.Write(filepath.Join(dir, "weird.xyz")); err != nil {
		t.Fatalf("Write unsupported ext: %v", err)
	}
	// Unwritable path.
	if err := im.Write(filepath.Join(dir, "nope", "out.png")); err == nil {
		t.Fatal("Write to missing directory should fail")
	}
}

func TestWriteToHelper(t *testing.T) {
	im := gradientImage(2, 2)
	// Encode failure branch.
	if err := writeTo(failWriter{}, im.rgba, "png"); err == nil {
		t.Fatal("writeTo with failing writer should error")
	}
	// Close failure branch (encode succeeds, Close fails).
	if err := writeTo(&failCloser{}, im.rgba, "png"); err == nil {
		t.Fatal("writeTo with failing closer should error")
	}
}

func TestSupportedAndFormatFromExt(t *testing.T) {
	if len(Supported()) != 6 || len(Encodable()) != 5 {
		t.Fatalf("format lists changed: %v / %v", Supported(), Encodable())
	}
	if _, err := formatFromExt("noext"); err == nil {
		t.Fatal("path without extension should error")
	}
	if _, err := formatFromExt("trailing."); err == nil {
		t.Fatal("trailing dot should error")
	}
	if _, err := formatFromExt("file.xyz"); err == nil {
		t.Fatal("unsupported extension should error")
	}
	if f, err := formatFromExt("PHOTO.JPG"); err != nil || f != "jpeg" {
		t.Fatalf("formatFromExt(JPG) = %q, %v", f, err)
	}
}
