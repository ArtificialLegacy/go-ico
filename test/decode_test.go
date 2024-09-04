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

func TestDecode_ValidICO(t *testing.T) {
	f, err := os.Open(TEST_FILE)
	if err != nil {
		t.Errorf("Error opening ICO file: %v", err)
	}

	data, err := goico.Decode(f)
	if err != nil {
		t.Errorf("Error decoding ICO file: %v", err)
	}

	if data.Header != goico.HEADER {
		t.Errorf("Header is not 0x%02x, got 0x%02x", goico.HEADER, data.Header)
	}

	if data.Type != goico.TYPE_ICO {
		t.Errorf("Type is not %d, got %d", goico.TYPE_ICO, data.Type)
	}

	if data.Count != 3 {
		t.Errorf("Count is not 3, got %d", data.Count)
	}

	if len(data.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(data.Entries))
	}

	validateEntry(t, 0, data.Entries[0], 16, 16, 0, 1, 32)
	validateEntry(t, 1, data.Entries[1], 32, 32, 0, 1, 32)
	validateEntry(t, 2, data.Entries[2], 48, 48, 0, 1, 32)

	// save images
	for i, entry := range data.Entries {
		f, err := os.Create(fmt.Sprintf("test_%d.bmp", i))
		if err != nil {
			t.Errorf("Error creating BMP file: %v", err)
		}
		defer f.Close()
		err = bmp.Encode(f, entry.Img)
		if err != nil {
			t.Errorf("Error encoding BMP file: %v", err)
		}

		// as png
		fp, err := os.Create(fmt.Sprintf("test_%d.png", i))
		if err != nil {
			t.Errorf("Error creating PNG file: %v", err)
		}
		defer fp.Close()
		err = png.Encode(fp, entry.Img)
		if err != nil {
			t.Errorf("Error encoding PNG file: %v", err)
		}
	}
}

func validateEntry(t *testing.T, i int, entry *goico.DIR_Entry, width, height, colors uint8, data1, data2 uint16) {
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

	if entry.Size == 0 {
		t.Errorf("entry %d size is 0", i)
	}

	if entry.Offset < goico.MIN_OFFSET {
		t.Errorf("entry %d offset is less than %d, got %d", i, goico.MIN_OFFSET, entry.Offset)
	}
}
