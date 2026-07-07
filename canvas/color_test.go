// Copyright (c) 2026, the go-ruby-images/images authors. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause license that can be
// found in the LICENSE file.

package canvas

import "testing"

func TestColorComponents(t *testing.T) {
	c := RGBA(0x11, 0x22, 0x33, 0x44)
	if c.R() != 0x11 || c.G() != 0x22 || c.B() != 0x33 || c.A() != 0x44 {
		t.Fatalf("components = %02x %02x %02x %02x", c.R(), c.G(), c.B(), c.A())
	}
	if RGB(1, 2, 3).A() != 0xff {
		t.Fatal("RGB should be opaque")
	}
	if c.Hex() != "#11223344" {
		t.Fatalf("Hex = %s", c.Hex())
	}
	r, g, b, a := White.RGBA()
	if r != 0xffff || g != 0xffff || b != 0xffff || a != 0xffff {
		t.Fatalf("White.RGBA = %d %d %d %d", r, g, b, a)
	}
}

func TestFromHex(t *testing.T) {
	cases := map[string]Color{
		"#fff":      RGBA(0xff, 0xff, 0xff, 0xff),
		"f00":       RGBA(0xff, 0x00, 0x00, 0xff),
		"#0f08":     RGBA(0x00, 0xff, 0x00, 0x88),
		"112233":    RGBA(0x11, 0x22, 0x33, 0xff),
		"#11223344": RGBA(0x11, 0x22, 0x33, 0x44),
		"#AABBCCDD": RGBA(0xaa, 0xbb, 0xcc, 0xdd), // uppercase digits
	}
	for in, want := range cases {
		got, err := FromHex(in)
		if err != nil || got != want {
			t.Fatalf("FromHex(%q) = %08x, %v; want %08x", in, uint32(got), err, uint32(want))
		}
	}
	// Bad-length and bad-digit inputs at each parser path.
	bad := []string{"", "12", "#12345", "ggg", "gggg", "gg2233", "1g2233", "gg223344", "1122g344"}
	for _, in := range bad {
		if _, err := FromHex(in); err == nil {
			t.Fatalf("FromHex(%q) should fail", in)
		}
	}
}

func TestCompose(t *testing.T) {
	// Opaque foreground wins outright.
	if got := Compose(Red, Blue); got != Red {
		t.Fatalf("opaque compose = %08x", uint32(got))
	}
	// Transparent foreground leaves the background.
	if got := Compose(Transparent, Blue); got != Blue {
		t.Fatalf("transparent compose = %08x", uint32(got))
	}
	// Partial alpha blends toward the foreground and yields a higher alpha
	// than either input pixel over an opaque background.
	half := RGBA(0xff, 0x00, 0x00, 0x80)
	out := Compose(half, RGBA(0x00, 0x00, 0xff, 0xff))
	if out.A() != 0xff {
		t.Fatalf("partial compose alpha = %02x", out.A())
	}
	if out.R() == 0 || out.B() == 0 {
		t.Fatalf("partial compose did not blend: %08x", uint32(out))
	}
	// Partial over a transparent background keeps the foreground alpha.
	over := Compose(half, Transparent)
	if over.A() != 0x80 {
		t.Fatalf("compose over transparent alpha = %02x", over.A())
	}
}
