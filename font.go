// Some of this code (namely the code for computing the
// width of a string in a given font) was copied from
// code.google.com/p/freetype-go/freetype/ which includes
// the following copyright notice:
// Copyright 2010 The Freetype-Go Authors. All rights reserved.
package main

import (
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

const (
	fontSize = 11
)

func calculateTextWidth(text string) (int, error) {
	binFont, _ := Asset("assets/DejaVuSans.ttf")
	font, err := truetype.Parse(binFont)
	if err != nil {
		return 0, err
	}

	scale := fontSize / float64(font.FUnitsPerEm())

	width := 0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range text {
		fUnitsPerEm := fixed.Int26_6(font.FUnitsPerEm())
		index := font.Index(rune)
		if hasPrev {
			width += int(font.Kern(fUnitsPerEm, prev, index))
		}
		width += int(font.HMetric(fUnitsPerEm, index).AdvanceWidth)
		prev, hasPrev = index, true
	}

	return int(float64(width) * scale), nil
}
