<p align="center"><img src="https://go-ruby-images.github.io/brand/social/go-ruby-images-images.png" alt="go-ruby-images/images" width="720"></p>

# images — go-ruby-images

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-images.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) image library** that unifies three Ruby image toolkits
behind one clean, Ruby-bindable Go API — with **no ImageMagick, no libvips, no
C QR/codec library**. It is an image backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) (rbgo), destined to
be exposed as `require "images"`, but is a **standalone, reusable** module — a
sibling of
[go-ruby-rqrcode](https://github.com/go-ruby-rqrcode/rqrcode),
[go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb/erb).

It combines, in one coherent module:

1. **Processing** — a MiniMagick / ruby-vips replacement: decode PNG, JPEG, GIF,
   BMP, TIFF and WebP, report dimensions / format / metadata, and `resize`,
   `thumbnail`, `crop`, `rotate`, `flip`, `flop`, `grayscale`, format `convert`
   and `to_blob` / `write` — over the standard library and `golang.org/x/image`.
2. **A chunky_png-style canvas** (sub-package `canvas`) — per-pixel access,
   primitive drawing (points, lines, rectangles, circles), alpha compositing,
   colour helpers (packed `0xRRGGBBAA`, hex parse), and PNG read/write.
3. **scikit-image operations** — blur, sharpen, unsharp mask, median, Sobel /
   Prewitt / Scharr / Laplacian / Canny edges, threshold / Otsu, morphology
   (erode / dilate / open / close) and colour ops, reusing the
   [go-images](https://github.com/go-images/images) engine rather than
   reimplementing it.

Every operation is **deterministic** (fixed algorithms, no randomness), so
results reproduce across the six supported 64-bit architectures and the
`js/wasm` and `wasip1/wasm` targets.

## Install

```sh
go get github.com/go-ruby-images/images
```

## Usage

### Processing (MiniMagick / ruby-vips style)

```go
import "github.com/go-ruby-images/images"

img, _ := images.Open("photo.jpg")
w, h := img.Dimensions()          // dimensions
_ = img.Format()                  // "jpeg"
_ = img.Metadata()                // {Width, Height, Format, ColorModel}

thumb, _ := img.Thumbnail(200, 200)     // fit inside a box, keep aspect
crop, _  := img.Crop(10, 10, 100, 100)  // region
rot, _   := img.Rotate(90)              // multiples of 90
flipped  := img.Flip().Flop()           // vertical then horizontal mirror

png, _ := img.Convert("png")
blob, _ := png.ToBlob()                 // encoded bytes
_ = png.Write("out.png")                // format from the path extension
```

### chunky_png-style canvas

```go
import "github.com/go-ruby-images/images/canvas"

c := canvas.New(64, 64, canvas.White)
_ = c.Set(1, 1, canvas.RGB(255, 0, 0))
c.Line(0, 0, 63, 63, canvas.Black)
c.Rect(8, 8, 40, 40, canvas.Black, canvas.Blue)   // stroke, fill
c.Circle(32, 32, 16, canvas.Green)
red, _ := canvas.FromHex("#ff0000")               // packed 0xRRGGBBAA
_ = c.SavePNG("draw.png")
```

### scikit-image operations

```go
img, _ := images.Open("scan.png")
edges, _ := img.Canny(1.0, 30, 90)     // Canny edges
blurred, _ := img.GaussianBlur(1.5)
bw := img.Otsu()                        // automatic threshold
opened, _ := img.Open(2)                // morphological opening
_ = img.Sobel(); _ = img.Sharpen(); _ = img.Invert()
```

## Tests & coverage

The suite is deterministic and self-contained (images are synthesised in
memory), and holds **100% statement coverage including every error branch**:

```sh
GOWORK=off go test ./... -covermode=count -coverprofile=cover.out
go tool cover -func=cover.out | grep -v 100.0%   # only the total remains
```

CI additionally builds and tests on the six supported 64-bit architectures
(amd64, arm64, riscv64, loong64, ppc64le, s390x), builds `js/wasm` and
`wasip1/wasm`, and runs on Linux, macOS and Windows.

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) 2026, the
go-ruby-images/images authors.
