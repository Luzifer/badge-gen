package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEmbeddedFontHash(t *testing.T) {
	// Check the embedded font did not change
	font, err := Asset("assets/DejaVuSans.ttf")
	if err != nil {
		t.Errorf("Could not load embedded font: %s", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(font))
	if hash != "3fdf69cabf06049ea70a00b5919340e2ce1e6d02b0cc3c4b44fb6801bd1e0d22" {
		t.Errorf("Embedded font changed: %s", hash)
	}
}

func TestStringLength(t *testing.T) {
	// As the font is embedded into the source the length calculation should not change
	w, err := calculateTextWidth("Test 123 öäüß … !@#%&")
	if err != nil {
		t.Errorf("Text length errored: %s", err)
	}
	if w != 138 {
		t.Errorf("Text length changed and is now %d", w)
	}
}
