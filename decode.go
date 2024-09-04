package goico

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"

	"golang.org/x/image/bmp"
)

type DIR struct {
	Header uint16
	Type   ICOType
	Count  uint16

	Entries []*DIR_Entry // length equal to Count
}

type DIR_Entry struct {
	Width  uint8
	Height uint8
	Colors uint8
	Data1  uint16 // ICO: color planes;   CUR: Hotspot X
	Data2  uint16 // ICO: bits per pixel; CUR: Hotspot Y
	Size   uint32
	Offset uint32

	Img image.Image
}

type FormatError string

func (e FormatError) Error() string { return "ico: invalid format: " + string(e) }

func NewFormatError(err string) FormatError {
	return FormatError(err)
}

func NewFormatErrorf(format string, args ...interface{}) FormatError {
	return FormatError(fmt.Sprintf(format, args...))
}

const EXT_ICO string = ".ico"
const EXT_CUR string = ".cur"

const HEADER uint16 = 0x00 // starts with 2 0 bytes.
const HEADER_LEN = 6
const ENTRY_LEN = 16
const MIN_OFFSET = 54

type ICOType uint16

const (
	TYPE_NIL ICOType = 0x00
	TYPE_ICO ICOType = 0x01 // .ico is type 1.
	TYPE_CUR ICOType = 0x02 // .cur is type 2.
)

func Decode(r io.ReadSeeker) (*DIR, error) {
	data := &DIR{}

	hv, tv, cv, err := decodeICONDIR(r)
	if err != nil {
		return nil, err
	}
	data.Header = hv
	data.Type = tv
	data.Count = cv

	entries, err := decodeICONDIRENTRY(r, cv)
	if err != nil {
		return nil, err
	}
	data.Entries = *entries

	for i := range cv {
		err := decodeICONDIRENTRYData(r, data.Entries[i])
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func decodeICONDIR(r io.Reader) (uint16, ICOType, uint16, error) {
	header := make([]byte, HEADER_LEN)
	n, err := r.Read(header)
	if err != nil {
		return 0, TYPE_NIL, 0, NewFormatErrorf("error reading header: %v", err)
	}
	if n != HEADER_LEN {
		return 0, TYPE_NIL, 0, NewFormatErrorf("error reading header: expected %d bytes, got %d", HEADER_LEN, n)
	}

	headerValue := binary.LittleEndian.Uint16(header[:2])
	if headerValue != HEADER {
		return 0, TYPE_NIL, 0, NewFormatErrorf("invalid header: expected 0x%02x, got 0x%02x", HEADER, headerValue)
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

func decodeICONDIRENTRY(r io.Reader, count uint16) (*[]*DIR_Entry, error) {
	entries := make([]*DIR_Entry, count)

	for i := range count {
		entry := make([]byte, ENTRY_LEN)
		n, err := r.Read(entry)
		if err != nil {
			return nil, NewFormatErrorf("error reading entry: %v", err)
		}
		if n != ENTRY_LEN {
			return nil, NewFormatErrorf("error reading entry: expected %d bytes, got %d", ENTRY_LEN, n)
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
		if offset < MIN_OFFSET {
			return nil, NewFormatErrorf("invalid offset: expected >= %d, got %d", MIN_OFFSET, offset)
		}

		entries[i] = &DIR_Entry{
			Width:  width,
			Height: height,
			Colors: colors,
			Data1:  data1,
			Data2:  data2,
			Size:   size,
			Offset: offset,
		}
	}

	return &entries, nil
}

type BMPHeader struct {
	Field     uint16
	Size      uint32
	Reserved1 uint16
	Reserved2 uint16
	Offset    uint32
}

func (h *BMPHeader) Bytes() []byte {
	b := make([]byte, 14)
	binary.LittleEndian.PutUint16(b[:2], h.Field)
	binary.LittleEndian.PutUint32(b[2:6], h.Size)
	binary.LittleEndian.PutUint16(b[6:8], h.Reserved1)
	binary.LittleEndian.PutUint16(b[8:10], h.Reserved2)
	binary.LittleEndian.PutUint32(b[10:14], h.Offset)
	return b
}

func decodeICONDIRENTRYData(r io.ReadSeeker, entry *DIR_Entry) error {
	p, err := r.Seek(int64(entry.Offset), io.SeekStart)
	if err != nil {
		return NewFormatErrorf("error seeking to entry data: %v", err)
	}
	if p != int64(entry.Offset) {
		return NewFormatErrorf("error seeking to entry data: expected offset %d, got %d", entry.Offset, p)
	}

	data := make([]byte, entry.Size)
	n, err := r.Read(data)
	if err != nil {
		return NewFormatErrorf("error reading entry data: %v", err)
	}
	if n != int(entry.Size) {
		return NewFormatErrorf("error reading entry data: expected %d bytes, got %d", entry.Size, n)
	}

	headerLength := binary.LittleEndian.Uint32(data[0:4])

	header := BMPHeader{
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
		return NewFormatErrorf("error decoding BMP data: %v", err)
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

	entry.Img = img
	return nil
}
