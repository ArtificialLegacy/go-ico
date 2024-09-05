package goico

import "image"

type ICOConfig struct {
	Count   int
	Type    ICOType
	Entries []ICOConfigEntry
	Largest int // index of the largest image
}

type ICOConfigEntry struct {
	Width  int
	Height int
	Colors int
	Data1  int // CUR: Hotspot X
	Data2  int // CUR: Hotspot Y
}

const EXT_ICO string = ".ico"
const EXT_CUR string = ".cur"

type ICOType uint16

const (
	TYPE_NIL ICOType = 0x00
	TYPE_ICO ICOType = 0x01 // .ico is type 1.
	TYPE_CUR ICOType = 0x02 // .cur is type 2.
)

func NewICOConfig(imgs []image.Image) (*ICOConfig, error) {
	count := len(imgs)
	entries := make([]ICOConfigEntry, count)

	for i, img := range imgs {
		width := img.Bounds().Dx()
		height := img.Bounds().Dy()

		if width > 256 || height > 256 {
			return nil, NewFormatErrorf("image %d is too large: %dx%d, cannot exceed 256", i, width, height)
		}

		entries[i] = ICOConfigEntry{
			Width:  width,
			Height: height,
		}
	}

	return &ICOConfig{
		Count:   count,
		Type:    TYPE_ICO,
		Entries: entries,
	}, nil
}

// hotspots is a list of x, y pairs. e.g. [x1, y1, x2, y2, ...]
func NewCURConfig(imgs []image.Image, hotspots []int) (*ICOConfig, error) {
	count := len(imgs)
	entries := make([]ICOConfigEntry, count)

	if len(hotspots) != count*2 {
		return nil, NewFormatErrorf("hotspots must be twice the number of images, got %d", len(hotspots))
	}

	for i, img := range imgs {
		width := img.Bounds().Dx()
		height := img.Bounds().Dy()

		if width > 256 || height > 256 {
			return nil, NewFormatErrorf("image %d is too large: %dx%d, cannot exceed 256", i, width, height)
		}

		entries[i] = ICOConfigEntry{
			Width:  width,
			Height: height,
			Data1:  hotspots[i*2],
			Data2:  hotspots[i*2+1],
		}
	}

	return &ICOConfig{
		Count:   count,
		Type:    TYPE_CUR,
		Entries: entries,
	}, nil
}

const headerV uint16 = 0x00 // starts with 2 0 bytes.
const header_len = 6
const entry_len = 16
