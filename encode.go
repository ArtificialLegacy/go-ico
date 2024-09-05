package goico

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"

	"golang.org/x/image/bmp"
)

type EncodeError string

func (e EncodeError) Error() string { return "ico: encoding error: " + string(e) }

func NewEncodeError(err string) EncodeError {
	return EncodeError(err)
}

func NewEncodeErrorf(format string, args ...interface{}) EncodeError {
	return EncodeError(fmt.Sprintf(format, args...))
}

func Encode(w io.WriteSeeker, ico *ICOConfig, imgs []image.Image) error {
	if ico.Type != TYPE_ICO && ico.Type != TYPE_CUR {
		return NewEncodeErrorf("invalid type: %d", ico.Type)
	}

	if ico.Count != len(imgs) {
		return NewEncodeErrorf("expected %d images, got %d", ico.Count, len(imgs))
	}

	if err := encodeICONDIR(w, ico.Type, uint16(ico.Count)); err != nil {
		return err
	}

	entryData := make([][]byte, ico.Count)
	entryPlanes := make([]uint16, ico.Count)
	entryBPP := make([]uint16, ico.Count)

	for i, img := range imgs {
		b, planes, bpp, err := encodeICONENTRYData(img)
		if err != nil {
			return err
		}

		entryData[i] = b
		entryPlanes[i] = planes
		entryBPP[i] = bpp
	}

	offset := 6 + entry_len*uint32(ico.Count) // offset for the start of the image data

	for i, entry := range ico.Entries {
		if ico.Type == TYPE_ICO {
			entry.Data1 = int(entryPlanes[i])
			entry.Data2 = int(entryBPP[i])
		}
		if err := encodeICONDIRENTRY(w, entry, entryData[i], offset); err != nil {
			return err
		}

		offset += uint32(len(entryData[i]))
	}

	return nil
}

func encodeICONDIR(w io.Writer, t ICOType, count uint16) error {
	header := make([]byte, header_len)
	binary.LittleEndian.PutUint16(header[0:2], headerV)
	binary.LittleEndian.PutUint16(header[2:4], uint16(t))
	binary.LittleEndian.PutUint16(header[4:6], count)

	n, err := w.Write(header)
	if err != nil {
		return err
	}
	if n != header_len {
		return NewFormatErrorf("error writing header: expected %d bytes, got %d", header_len, n)
	}

	return nil
}

func encodeICONDIRENTRY(w io.WriteSeeker, configEntry ICOConfigEntry, data []byte, offset uint32) error {
	width := configEntry.Width
	if width == 256 {
		width = 0
	}
	height := configEntry.Height
	if height == 256 {
		height = 0
	}

	entry := make([]byte, entry_len)
	entry[0] = uint8(width)
	entry[1] = uint8(height)
	entry[2] = uint8(configEntry.Colors)
	entry[3] = 0 // reserved

	binary.LittleEndian.PutUint16(entry[4:6], uint16(configEntry.Data1))
	binary.LittleEndian.PutUint16(entry[6:8], uint16(configEntry.Data2))

	binary.LittleEndian.PutUint32(entry[8:12], uint32(len(data)))
	binary.LittleEndian.PutUint32(entry[12:16], offset)

	n, err := w.Write(entry)
	if err != nil {
		return err
	}
	if n != entry_len {
		return NewEncodeErrorf("error writing entry: expected %d bytes, got %d", entry_len, n)
	}

	pos, _ := w.Seek(0, io.SeekCurrent)
	_, err = w.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return NewEncodeErrorf("error seeking to entry data: %v", err)
	}
	n, err = w.Write(data)
	if err != nil {
		return NewEncodeErrorf("error writing entry data: %v", err)
	}
	if n != len(data) {
		return NewEncodeErrorf("error writing entry data: expected %d bytes, got %d", len(data), n)
	}
	w.Seek(pos, io.SeekStart)

	return nil
}

func encodeICONENTRYData(img image.Image) ([]byte, uint16, uint16, error) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	if width > 256 || height > 256 {
		return nil, 0, 0, NewEncodeErrorf("image dimensions too large: %dx%d, favicon cannot exceed 256 in either dimension", width, height)
	}

	var w bytes.Buffer
	err := bmp.Encode(&w, img)
	if err != nil {
		return nil, 0, 0, NewEncodeErrorf("error encoding BMP: %v", err)
	}

	b := w.Bytes()
	b = b[14:] // remove BMP header

	binary.LittleEndian.PutUint32(b[8:12], uint32(img.Bounds().Dy()*2)) // ico requires 2x height

	bpp := binary.LittleEndian.Uint16(b[14:16]) // grab bpp for setting entry data
	return b, 1, bpp, nil
}
