package swf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// ExtractDoABCs extracts DoABC tag payloads from a SWF byte slice (supports FWS and CWS)
func ExtractDoABCs(swf []byte) ([][]byte, error) {
	if len(swf) < 8 { return nil, errors.New("swf too short") }
	sig := string(swf[0:3])
	if sig != "FWS" && sig != "CWS" && sig != "ZWS" { return nil, fmt.Errorf("unsupported signature %s", sig) }
	var payload []byte
	switch sig {
	case "FWS": payload = swf[8:]
	case "CWS":
		r := bytes.NewReader(swf[8:])
		rz, err := zlib.NewReader(r)
		if err != nil { return nil, err }
		defer rz.Close()
		un, err := io.ReadAll(rz); if err != nil { return nil, err }
		payload = un
	case "ZWS": return nil, errors.New("ZWS not supported")
	}

	r := bytes.NewReader(payload)
	// skip RECT (parse minimal)
	var first byte
	if err := binary.Read(r, binary.BigEndian, &first); err != nil { return nil, err }
	nbits := int(first >> 3)
	rectBits := 5 + 4*nbits
	rectBytes := (rectBits + 7) / 8
	if rectBytes > 1 { if _, err := r.Seek(int64(rectBytes-1), io.SeekCurrent); err != nil { return nil, err } }
	// skip frameRate(2) frameCount(2)
	if _, err := r.Seek(4, io.SeekCurrent); err != nil { return nil, err }

	var abcs [][]byte
	for {
		var tagHdr uint16
		if err := binary.Read(r, binary.LittleEndian, &tagHdr); err != nil { if err == io.EOF { break }; return nil, err }
		tagCode := int(tagHdr >> 6)
		tagLen := int(tagHdr & 0x3F)
		if tagLen == 0x3F {
			var longLen uint32
			if err := binary.Read(r, binary.LittleEndian, &longLen); err != nil { return nil, err }
			tagLen = int(longLen)
		}
		payload := make([]byte, tagLen)
		if tagLen > 0 { if _, err := io.ReadFull(r, payload); err != nil { return nil, err } }
		if tagCode == 82 {
			buf := bytes.NewReader(payload)
			var flags uint32
			_ = binary.Read(buf, binary.LittleEndian, &flags)
			// read name until NUL
			for {
				b, err := buf.ReadByte(); if err != nil { break }
				if b == 0 { break }
			}
			abcBytes, _ := io.ReadAll(buf)
			abcs = append(abcs, abcBytes)
		}
		if tagCode == 0 { break }
	}
	return abcs, nil
}
