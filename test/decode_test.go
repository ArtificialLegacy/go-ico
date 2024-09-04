package test

import (
	"os"
	"testing"

	goico "github.com/ArtificialLegacy/go-ico"
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
}
