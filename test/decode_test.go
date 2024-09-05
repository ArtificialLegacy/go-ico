package test

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	goico "github.com/ArtificialLegacy/go-ico"
	"golang.org/x/image/bmp"
)

const TEST_FILE = "favicon.ico"

func TestDecodeConfig(t *testing.T) {
	f, err := os.Open(TEST_FILE)
	if err != nil {
		t.Errorf("Error opening ICO file: %v", err)
	}
	defer f.Close()

	config, err := goico.DecodeConfig(f)
	if err != nil {
		t.Errorf("Error decoding ICO file: %v", err)
	}

	if config.Count != 3 {
		t.Errorf("Count is not 3, got %d", config.Count)
	}

	if config.Type != goico.TYPE_ICO {
		t.Errorf("Type is not %d, got %d", goico.TYPE_ICO, config.Type)
	}

	if len(config.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(config.Entries))
	}

	if config.Largest != 2 {
		t.Errorf("Largest is not 2, got %d", config.Largest)
	}

	for i, entry := range config.Entries {
		if entry.Width != 16 && entry.Width != 32 && entry.Width != 48 {
			t.Errorf("entry %d width is not 16, 32, or 48, got %d", i, entry.Width)
		}
		if entry.Height != 16 && entry.Height != 32 && entry.Height != 48 {
			t.Errorf("entry %d height is not 16, 32, or 48, got %d", i, entry.Height)
		}

		if entry.Colors != 0 {
			t.Errorf("entry %d colors is not 0, got %d", i, entry.Colors)
		}

		if entry.Data1 != 0 {
			t.Errorf("entry %d data1 is not 0, got %d", i, entry.Data1)
		}

		if entry.Data2 != 0 {
			t.Errorf("entry %d data2 is not 0, got %d", i, entry.Data2)
		}
	}
}

func TestDecode_ValidICO(t *testing.T) {
	f, err := os.Open(TEST_FILE)
	if err != nil {
		t.Errorf("Error opening ICO file: %v", err)
	}
	defer f.Close()

	config, imgs, err := goico.Decode(f)
	if err != nil {
		t.Errorf("Error decoding ICO file: %v", err)
	}

	if config.Type != goico.TYPE_ICO {
		t.Errorf("Type is not %d, got %d", goico.TYPE_ICO, config.Type)
	}

	if config.Count != 3 {
		t.Errorf("Count is not 3, got %d", config.Count)
	}

	if len(config.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(config.Entries))
	}

	validateEntry(t, 0, config.Entries[0], 16, 16, 0, 0, 0)
	validateEntry(t, 1, config.Entries[1], 32, 32, 0, 0, 0)
	validateEntry(t, 2, config.Entries[2], 48, 48, 0, 0, 0)

	// save images
	for i, img := range imgs {
		f, err := os.Create(fmt.Sprintf("test_%d.bmp", i))
		if err != nil {
			t.Errorf("Error creating BMP file: %v", err)
		}
		defer f.Close()
		err = bmp.Encode(f, img)
		if err != nil {
			t.Errorf("Error encoding BMP file: %v", err)
		}

		// as png
		fp, err := os.Create(fmt.Sprintf("test_%d.png", i))
		if err != nil {
			t.Errorf("Error creating PNG file: %v", err)
		}
		defer fp.Close()
		err = png.Encode(fp, img)
		if err != nil {
			t.Errorf("Error encoding PNG file: %v", err)
		}
	}
}

func validateEntry(t *testing.T, i int, entry goico.ICOConfigEntry, width, height, colors, data1, data2 int) {
	if entry.Width != width {
		t.Errorf("entry %d width is not %d, got %d", i, width, entry.Width)
	}

	if entry.Height != height {
		t.Errorf("entry %d height is not %d, got %d", i, height, entry.Height)
	}

	if entry.Colors != colors {
		t.Errorf("entry %d colors is not %d, got %d", i, colors, entry.Colors)
	}

	if entry.Data1 != data1 {
		t.Errorf("entry %d data1 is not %d, got %d", i, data1, entry.Data1)
	}

	if entry.Data2 != data2 {
		t.Errorf("entry %d data2 is not %d, got %d", i, data2, entry.Data2)
	}
}
