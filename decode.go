package goico

import (
	"encoding/binary"
	"fmt"
	"io"
)

type DIR struct {
	Header uint16
	Type   ICOType
	Count  uint16

	Entries []Dir_Entry // length equal to Count
}

type Dir_Entry interface {
	Type() ICOType
}

type DIR_EntryICO struct {
	Width        uint8
	Height       uint8
	Colors       uint8
	ColorPlanes  uint16
	BitsPerPixel uint16
	Size         uint32
	Offset       uint32
}

func (e *DIR_EntryICO) Type() ICOType { return TYPE_ICO }

type DIR_EntryCUR struct {
	Width    uint8
	Height   uint8
	Colors   uint8
	HotspotX uint16
	HotspotY uint16
	Size     uint32
	Offset   uint32
}

func (e *DIR_EntryCUR) Type() ICOType { return TYPE_CUR }

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

type ICOType uint16

const (
	TYPE_NIL ICOType = 0x00
	TYPE_ICO ICOType = 0x01 // .ico is type 1.
	TYPE_CUR ICOType = 0x02 // .cur is type 2.
)

func Decode(r io.Reader) (*DIR, error) {
	data := &DIR{}

	hv, tv, cv, err := decodeICONDIR(r)
	if err != nil {
		return nil, err
	}
	data.Header = hv
	data.Type = tv
	data.Count = cv

	return data, nil
}

func decodeICONDIR(r io.Reader) (uint16, ICOType, uint16, error) {
	var n int
	var err error

	header := make([]byte, HEADER_LEN)
	n, err = r.Read(header)
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
