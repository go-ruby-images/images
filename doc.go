// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

// Package images is a pure-Go (no cgo) image library that backs the Ruby
// `images` gem for the go-embedded-ruby interpreter (rbgo). It unifies three
// capability sets behind one clean, Ruby-bindable Go API:
//
//   - Processing (MiniMagick / ruby-vips replacement): the [Image] type decodes
//     PNG, JPEG, GIF, BMP, TIFF and WebP, reports dimensions/format/metadata,
//     and offers resize, thumbnail, crop, rotate, flip, flop, grayscale, format
//     conversion, and blob/file output — with no ImageMagick or libvips.
//   - A chunky_png-style canvas lives in the sub-package
//     github.com/go-ruby-images/images/canvas, with per-pixel access, primitive
//     drawing, alpha composition and PNG I/O.
//   - scikit-image / go-images operations are exposed as methods on [Image]
//     (blur, sharpen, edges, threshold, morphology, colour ops), reusing the
//     github.com/go-images/images engine rather than reimplementing it.
//
// Every operation is deterministic (fixed algorithms, no randomness) so the
// results are reproducible across the six supported 64-bit architectures and
// the js/wasm and wasip1/wasm targets.
package images
