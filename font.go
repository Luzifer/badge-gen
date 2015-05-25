// Some of this code (namely the code for computing the
// width of a string in a given font) was copied from
// code.google.com/p/freetype-go/freetype/ which includes
// the following copyright notice:
// Copyright 2010 The Freetype-Go Authors. All rights reserved.
package main

import "code.google.com/p/freetype-go/freetype/truetype"

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
		index := font.Index(rune)
		if hasPrev {
			width += int(font.Kerning(font.FUnitsPerEm(), prev, index))
		}
		width += int(font.HMetric(font.FUnitsPerEm(), index).AdvanceWidth)
		prev, hasPrev = index, true
	}

	return int(float64(width) * scale), nil
}
