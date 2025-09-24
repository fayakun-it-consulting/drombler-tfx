package abc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// ABCFile represents a parsed ABC file (simplified for this release).
type ABCFile struct {
	MinorVersion uint16
	MajorVersion uint16
	Strings      []string
	Methods      []MethodInfo
	// Placeholder maps for future trait/multiname parsing
}

// MethodInfo is a very small representation with raw code bytes.
type MethodInfo struct {
	Code []byte
}

// readU30 reads a variable-length U30 value used in ABC.
func readU30(r *bytes.Reader) (uint32, error) {
	var result uint32
	var shift uint
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift >= 35 {
			return 0, fmt.Errorf("invalid u30")		}
	}
	return result, nil
}

// ParseABC parses a minimal ABC file: header (minor/major), string pool, methods list with code bytes.
// This parser matches the toy ABC examples included here. Full ABC parsing remains future work.
func ParseABC(data []byte) (*ABCFile, error) {
	r := bytes.NewReader(data)

	var minor, major uint16
	if err := binary.Read(r, binary.LittleEndian, &minor); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &major); err != nil {
		return nil, err
	}

	af := &ABCFile{MinorVersion: minor, MajorVersion: major}

	// strings pool
	strCount, err := readU30(r)
	if err != nil {
		return nil, err
	}
	if strCount == 0 {
		strCount = 1
	}
	af.Strings = make([]string, strCount)
	for i := uint32(1); i < strCount; i++ {
		ln, err := readU30(r)
		if err != nil {
			return nil, err
		}
		buf := make([]byte, ln)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		af.Strings[i] = string(buf)
	}

	// methods
	methodCount, err := readU30(r)
	if err != nil {
		return nil, err
	}
	af.Methods = make([]MethodInfo, methodCount)
	for i := uint32(0); i < methodCount; i++ {
		codeLen, err := readU30(r)
		if err != nil {
			return nil, err
		}
		code := make([]byte, codeLen)
		if _, err := io.ReadFull(r, code); err != nil {
			return nil, err
		}
		af.Methods[i] = MethodInfo{Code: code}
	}

	return af, nil
}
