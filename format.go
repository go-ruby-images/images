// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package images

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	// Register the additional decoders so image.Decode / image.DecodeConfig
	// can sniff BMP, TIFF and WebP in addition to the standard-library PNG,
	// JPEG and GIF formats. WebP is decode-only in pure Go.
	_ "golang.org/x/image/webp"
)

// jpegQuality is the quality used when (re-)encoding JPEG output. It matches
// the go-images sibling so round-trips are consistent across the ecosystem.
const jpegQuality = 90

// Supported reports the canonical format names this package can decode.
func Supported() []string {
	return []string{"png", "jpeg", "gif", "bmp", "tiff", "webp"}
}

// Encodable reports the canonical format names this package can encode. WebP is
// intentionally absent: there is no pure-Go WebP encoder, so encoding to WebP
// returns an error rather than pulling in a cgo dependency.
func Encodable() []string {
	return []string{"png", "jpeg", "gif", "bmp", "tiff"}
}

// normalizeFormat lower-cases a format name and folds common aliases ("jpg" to
// "jpeg", "tif" to "tiff") onto their canonical form.
func normalizeFormat(name string) string {
	switch f := strings.ToLower(strings.TrimSpace(name)); f {
	case "jpg":
		return "jpeg"
	case "tif":
		return "tiff"
	default:
		return f
	}
}

// decode reads an image from r, converting it to *image.RGBA and returning the
// detected canonical format name. It reports an error for unrecognised or
// corrupt data.
func decode(r io.Reader) (*image.RGBA, string, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, "", fmt.Errorf("images: decode: %w", err)
	}
	return toRGBA(img), normalizeFormat(format), nil
}

// encode writes img to w in the named format. It returns an error for WebP
// (unsupported for encoding) and for any unknown format.
func encode(w io.Writer, img image.Image, format string) error {
	switch normalizeFormat(format) {
	case "png":
		if err := png.Encode(w, img); err != nil {
			return fmt.Errorf("images: encode png: %w", err)
		}
	case "jpeg":
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return fmt.Errorf("images: encode jpeg: %w", err)
		}
	case "gif":
		if err := gif.Encode(w, img, nil); err != nil {
			return fmt.Errorf("images: encode gif: %w", err)
		}
	case "bmp":
		if err := bmp.Encode(w, img); err != nil {
			return fmt.Errorf("images: encode bmp: %w", err)
		}
	case "tiff":
		if err := tiff.Encode(w, img, nil); err != nil {
			return fmt.Errorf("images: encode tiff: %w", err)
		}
	case "webp":
		return fmt.Errorf("images: encode: webp encoding is not supported in pure Go")
	default:
		return fmt.Errorf("images: encode: unknown format %q", format)
	}
	return nil
}

// formatFromExt maps a file path's extension to a canonical format name.
func formatFromExt(path string) (string, error) {
	i := strings.LastIndexByte(path, '.')
	if i < 0 || i == len(path)-1 {
		return "", fmt.Errorf("images: cannot infer format from path %q", path)
	}
	ext := normalizeFormat(path[i+1:])
	for _, f := range Supported() {
		if f == ext {
			return ext, nil
		}
	}
	return "", fmt.Errorf("images: unsupported file extension %q", path[i:])
}
