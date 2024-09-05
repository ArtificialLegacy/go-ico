package goico

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"

	"golang.org/x/image/bmp"
)

type dirEntry struct {
	Width  uint8
	Height uint8
	Colors uint8
	Data1  uint16 // ICO: color planes;   CUR: Hotspot X
	Data2  uint16 // ICO: bits per pixel; CUR: Hotspot Y
	Size   uint32
	Offset uint32
}

type FormatError string

func (e FormatError) Error() string { return "ico: invalid format: " + string(e) }

func NewFormatError(err string) FormatError {
	return FormatError(err)
}

func NewFormatErrorf(format string, args ...interface{}) FormatError {
	return FormatError(fmt.Sprintf(format, args...))
}

func Decode(r io.ReadSeeker) (*ICOConfig, []image.Image, error) {
	_, tv, cv, err := decodeICONDIR(r)
	if err != nil {
		return nil, []image.Image{}, err
	}

	data := &ICOConfig{
		Count:   int(cv),
		Type:    tv,
		Entries: []ICOConfigEntry{},
	}

	imgs := []image.Image{}

	largest := 0
	for i := range int(cv) {
		entryData, err := decodeICONDIRENTRY(r)
		if err != nil {
			return nil, []image.Image{}, err
		}

		// size of 0 is used for a size of 256
		width := int(entryData.Width)
		if width == 0 {
			width = 256
		}
		height := int(entryData.Height)
		if height == 0 {
			height = 256
		}

		size := width * height
		if size > largest {
			largest = size
			data.Largest = i
		}

		// CUR: Hotspot X and Y are stored in Data1 and Data2.
		// Remove them for ICO, as they are not used after decoding.
		data1 := int(entryData.Data1)
		data2 := int(entryData.Data2)
		if tv == TYPE_ICO {
			data1 = 0
			data2 = 0
		}

		data.Entries = append(data.Entries, ICOConfigEntry{
			Width:  width,
			Height: height,
			Colors: int(entryData.Colors),
			Data1:  data1,
			Data2:  data2,
		})

		img, err := decodeICONDIRENTRYData(r, entryData)
		if err != nil {
			return nil, []image.Image{}, err
		}

		imgs = append(imgs, img)
	}

	return data, imgs, nil
}

func DecodeConfig(r io.Reader) (*ICOConfig, error) {
	_, tv, cv, err := decodeICONDIR(r)
	if err != nil {
		return nil, err
	}

	data := &ICOConfig{
		Count:   int(cv),
		Type:    tv,
		Entries: []ICOConfigEntry{},
	}

	largest := 0
	for i := range int(cv) {
		entryData, err := decodeICONDIRENTRY(r)
		if err != nil {
			return nil, err
		}

		// size of 0 is used for a size of 256
		width := int(entryData.Width)
		if width == 0 {
			width = 256
		}
		height := int(entryData.Height)
		if height == 0 {
			height = 256
		}

		size := width * height
		if size > largest {
			largest = size
			data.Largest = i
		}

		// CUR: Hotspot X and Y are stored in Data1 and Data2.
		// Remove them for ICO, as they are not used after decoding.
		data1 := int(entryData.Data1)
		data2 := int(entryData.Data2)
		if tv == TYPE_ICO {
			data1 = 0
			data2 = 0
		}

		data.Entries = append(data.Entries, ICOConfigEntry{
			Width:  width,
			Height: height,
			Colors: int(entryData.Colors),
			Data1:  data1,
			Data2:  data2,
		})
	}

	return data, nil
}

func decodeICONDIR(r io.Reader) (uint16, ICOType, uint16, error) {
	header := make([]byte, header_len)
	n, err := r.Read(header)
	if err != nil {
		return 0, TYPE_NIL, 0, NewFormatErrorf("error reading header: %v", err)
	}
	if n != header_len {
		return 0, TYPE_NIL, 0, NewFormatErrorf("error reading header: expected %d bytes, got %d", header_len, n)
	}

	headerValue := binary.LittleEndian.Uint16(header[:2])
	if headerValue != headerV {
		return 0, TYPE_NIL, 0, NewFormatErrorf("invalid header: expected 0x%02x, got 0x%02x", headerV, headerValue)
	}

	typeValue := ICOType(binary.LittleEndian.Uint16(header[2:4]))
	if typeValue != TYPE_ICO && typeValue != TYPE_CUR {
		return 0, TYPE_NIL, 0, NewFormatErrorf("invalid type: expected %d or %d, got %d", TYPE_ICO, TYPE_CUR, typeValue)
	}

	countValue := binary.LittleEndian.Uint16(header[4:6])
	if countValue == 0 {
		return 0, TYPE_NIL, 0, NewFormatError("invalid image count: expected > 0, got 0")
	}

	return headerValue, typeValue, countValue, nil
}

func decodeICONDIRENTRY(r io.Reader) (*dirEntry, error) {
	entry := make([]byte, entry_len)
	n, err := r.Read(entry)
	if err != nil {
		return nil, NewFormatErrorf("error reading entry: %v", err)
	}
	if n != entry_len {
		return nil, NewFormatErrorf("error reading entry: expected %d bytes, got %d", entry_len, n)
	}

	width := uint8(entry[0])
	height := uint8(entry[1])
	colors := uint8(entry[2])
	reserved := uint8(entry[3])
	if reserved != 0 {
		return nil, NewFormatErrorf("invalid reserved byte: expected 0, got %d", reserved)
	}

	data1 := binary.LittleEndian.Uint16(entry[4:6])
	data2 := binary.LittleEndian.Uint16(entry[6:8])
	size := binary.LittleEndian.Uint32(entry[8:12])
	if size == 0 {
		return nil, NewFormatError("invalid size: expected > 0, got 0")
	}

	offset := binary.LittleEndian.Uint32(entry[12:16])

	return &dirEntry{
		Width:  width,
		Height: height,
		Colors: colors,
		Data1:  data1,
		Data2:  data2,
		Size:   size,
		Offset: offset,
	}, nil
}

type bmpheader struct {
	Field     uint16
	Size      uint32
	Reserved1 uint16
	Reserved2 uint16
	Offset    uint32
}

func (h *bmpheader) Bytes() []byte {
	b := make([]byte, 14)
	binary.LittleEndian.PutUint16(b[:2], h.Field)
	binary.LittleEndian.PutUint32(b[2:6], h.Size)
	binary.LittleEndian.PutUint16(b[6:8], h.Reserved1)
	binary.LittleEndian.PutUint16(b[8:10], h.Reserved2)
	binary.LittleEndian.PutUint32(b[10:14], h.Offset)
	return b
}

func decodeICONDIRENTRYData(r io.ReadSeeker, entry *dirEntry) (image.Image, error) {
	currPos, _ := r.Seek(0, io.SeekCurrent)
	p, err := r.Seek(int64(entry.Offset), io.SeekStart)
	if err != nil {
		return nil, NewFormatErrorf("error seeking to entry data: %v", err)
	}
	if p != int64(entry.Offset) {
		return nil, NewFormatErrorf("error seeking to entry data: expected offset %d, got %d", entry.Offset, p)
	}

	data := make([]byte, entry.Size)
	n, err := r.Read(data)
	if err != nil {
		return nil, NewFormatErrorf("error reading entry data: %v", err)
	}
	if n != int(entry.Size) {
		return nil, NewFormatErrorf("error reading entry data: expected %d bytes, got %d", entry.Size, n)
	}

	headerLength := binary.LittleEndian.Uint32(data[0:4])
	if headerLength == 137 {
		return nil, NewFormatError("png format not supported")
	}

	header := bmpheader{
		Field:     0x4d42,          // "BM"
		Size:      entry.Size + 14, // size of BMP header + size of data
		Reserved1: 0,
		Reserved2: 0,
		Offset:    14 + headerLength,
	}

	data[8] = entry.Height
	data = append(header.Bytes(), data...)

	img, err := bmp.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, NewFormatErrorf("error decoding BMP data: %v", err)
	}

	// alpha hack
	if entry.Data2 == 32 {
		nrgba := img.(*image.NRGBA)

		y0 := int(entry.Height) - 1
		y1 := -1
		yDelta := -1
		if int(entry.Height) < 0 {
			y0 = 0
			y1 = int(entry.Height)
			yDelta = 1
		}

		width := int(entry.Width) * 4
		dIndex := header.Offset + 3

		for y := y0; y != y1; y += yDelta {
			p := nrgba.Pix[y*nrgba.Stride : y*nrgba.Stride+width]
			for i := 0; i < len(p); i += 4 {
				p[i+3] = data[dIndex]
				dIndex += 4
			}
		}
	}

	r.Seek(currPos, io.SeekStart)
	return img, nil
}
