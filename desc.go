//
// desc.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"fmt"
	"io"

	"github.com/OneOfOne/xxhash"
)

const (
	lzFLG             = 0
	lzBD              = 1
	lzMaxFrameDescLen = 11
)

const (
	lz64KB  = 65536
	lz256KB = 262144
	lz1MB   = 1048576
	lz4MB   = 4194304
)

type FrameDesc struct {
	Independent        bool
	HasBlockChecksum   bool
	HasContentChecksum bool
	BlockMaxSize       int
	ContentSize        uint64
}

func ReadFrameDesc(r io.Reader) (*FrameDesc, error) {
	var b [lzMaxFrameDescLen]byte
	// read FLG byte + BD byte + (HC byte or first byte of ContentSize)
	err := read(r, b[:3])
	if err != nil {
		return nil, err
	}
	// check version is 01
	if b[lzFLG]&0xc0 != 0x40 {
		return nil, fmt.Errorf("FrameDesc: version must be 01")
	}
	// check reserved bits are 0
	if b[lzFLG]&0x03 != 0 && b[lzBD]&0x8f != 0 {
		return nil, fmt.Errorf("FrameDesc: reserved bits must be zero")
	}
	bSize := lzBlockMaxSize(b[lzBD] & 0x70 >> 4)
	if bSize == -1 {
		return nil, fmt.Errorf("FrameDesc: unsupported Block Maximum size")
	}
	// optional ContentSize field
	var contentSize uint64
	i := lzBD + 1
	if b[lzFLG]&0x08 != 0 {
		// first ContentSize byte is already in b, read rest + HC byte
		err = read(r, b[i+1:i+9])
		if err != nil {
			return nil, err
		}
		contentSize = leUint64(b[i : i+8])
		i += 8
	}
	// HC byte
	hChecksum := byte((xxhash.Checksum32(b[:i]) >> 8) & 0xff)
	if hChecksum != b[i] {
		return nil, fmt.Errorf("FrameDesc: checksum mismatch %02x != %02x", hChecksum, b[i])
	}
	return &FrameDesc{
		Independent:        b[lzFLG]&0x20 != 0,
		HasBlockChecksum:   b[lzFLG]&0x10 != 0,
		HasContentChecksum: b[lzFLG]&0x04 != 0,
		BlockMaxSize:       bSize,
		ContentSize:        contentSize,
	}, nil
}

func lzBlockMaxSize(b byte) int {
	switch b {
	case 4:
		return lz64KB
	case 5:
		return lz256KB
	case 6:
		return lz1MB
	case 7:
		return lz4MB
	default:
		return -1
	}
}
