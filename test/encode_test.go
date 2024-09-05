package test

import (
	"image"
	"image/color"
	"os"
	"testing"

	goico "github.com/ArtificialLegacy/go-ico"
)

func TestEncode(t *testing.T) {
	imgs := []image.Image{
		createTestImage(16, 16, color.NRGBA{255, 0, 0, 128}),
		createTestImage(32, 32, color.NRGBA{0, 255, 0, 128}),
		createTestImage(48, 48, color.NRGBA{0, 0, 255, 128}),
	}

	ico, err := goico.NewICOConfig(imgs)
	if err != nil {
		t.Errorf("Error creating ICO config: %v", err)
	}

	f, err := os.Create("test.ico")
	if err != nil {
		t.Errorf("Error creating test ICO file: %v", err)
	}
	defer f.Close()

	err = goico.Encode(f, ico, imgs)
	if err != nil {
		t.Errorf("Error encoding ICO: %v", err)
	}
}

func createTestImage(width, height int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, c)
		}
	}

	return img
}
