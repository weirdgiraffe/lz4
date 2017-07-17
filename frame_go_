//
// frame.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"encoding/binary"
	"fmt"

	"github.com/OneOfOne/xxhash"
)

const lz4Magic uint32 = 0x184d2204 // little endian

type Frame struct {
	Desc  FrameDesc
	Block []byte
	End   [4]byte
}

func (f *Frame) Decode(in []byte) error {
	if len(in) < 4+3+2+4+4 {
		return fmt.Errorf("buffer is too small")
	}
	magic := binary.LittleEndian.Uint32(in)
	if magic != lz4Magic {
		return fmt.Errorf("magic not match")
	}
	err := f.Desc.Decode(in[4:])
	if err != nil {
		return err
	}
	n := 1 // decode block(...)
	chksum := binary.LittleEndian.Uint32(in)
	h := xxhash.Checksum32(in[:n])
	if h != chksum {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

type lzBlockSize int

const (
	lz64KB  lzBlockSize = 65536
	lz256KB             = 262144
	lz1MB               = 1048576
	lz4MB               = 4194304
)

type FrameDesc struct {
	Version            int
	Independent        bool
	HasBlockChksum     bool
	HasContentSize     bool
	HasContentChksum   bool
	HasChksum          bool
	BlockMaxSize       lzBlockSize
	DecodedContentSize uint64
}

func (d *FrameDesc) Decode(in []byte) error {
	if len(in) < 3 {
		return fmt.Errorf("Buffer size is < then minimal Descriptor size")
	}
	// FLG byte
	if in[0]&0xc0>>6 != 1 {
		return fmt.Errorf("Version must be 01 in FLG byte")
	}
	d.Independent = (in[0] & 0x20) != 0
	d.HasBlockChksum = (in[0] & 0x10) != 0
	d.HasContentSize = (in[0]&0x08 != 0)
	d.HasContentChksum = (in[0] & 0x40) != 0

	// BD byte
	if in[1]&0x8f != 0 {
		return fmt.Errorf("Reserved bits must be zero in BD byte")
	}
	switch (in[1] & 0x70) >> 4 {
	case 4:
		d.BlockMaxSize = lz64KB
	case 5:
		d.BlockMaxSize = lz256KB
	case 6:
		d.BlockMaxSize = lz1MB
	case 7:
		d.BlockMaxSize = lz4MB
	default:
		return fmt.Errorf("Block Max size should be 4,5,6,7 in BD byte")
	}

	chksumIndx := 2
	if d.HasContentSize {
		if len(in) < 11 {
			return fmt.Errorf("Buffer size is < then Descriptor size + Content size")
		}
		d.DecodedContentSize = binary.LittleEndian.Uint64(in[2:])
		chksumIndx = 10
	}
	h := byte((xxhash.Checksum32(in[:chksumIndx]) >> 8) & 0xff)
	if h != in[chksumIndx] {
		return fmt.Errorf("Checksum mismatch")
	}
	return nil
}

type Checksum struct {
}
