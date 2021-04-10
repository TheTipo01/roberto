package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func BenchmarkBestemmia(b *testing.B) {
	// Read adjectives if they are not loaded before
	if len(adjectives) == 0 {
		initializeAdjectives()
	}

	for i := 0; i < b.N; i++ {
		bestemmia()
	}
}

func TestInitializeAdjectives(t *testing.T) {
	initializeAdjectives()

	if len(adjectives) == 0 {
		t.Error("Adjectives slice is empty")
	}
}

func TestBestemmia(t *testing.T) {
	// Read adjectives if they are not loaded before
	if len(adjectives) == 0 {
		initializeAdjectives()
	}

	if strings.TrimSpace(bestemmia()) == "" {
		t.Error("Generated string is empty")
	}
}

func TestGenAudio(t *testing.T) {
	if runtime.GOOS == "windows" {
		_ = os.Remove("./temp/NP5M2VS4G9AQEEIPC6V6DQH5J1RGS4PE.dca")
		uuid := genAudio("AGAGAGAGAGA")
		stat, err := os.Stat("./temp/NP5M2VS4G9AQEEIPC6V6DQH5J1RGS4PE.dca")

		if uuid != "NP5M2VS4G9AQEEIPC6V6DQH5J1RGS4PE.dca" {
			t.Error("Hash mismatch")
		} else {
			if err != nil {
				t.Error("File doesn't exist")
			} else {
				if !(stat.Size() > 0) {
					t.Error("File is empty")
				}
			}
		}
	} else {
		t.Skip("Audio generation is only supported by windows")
	}

}
