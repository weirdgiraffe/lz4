//
// fdesc.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"io"

	"github.com/OneOfOne/xxhash"
)

type FrameDesc struct {
	Independent        bool
	HasBlockChecksum   bool
	HasContentChecksum bool
	HasDictID          bool
	BlockMaxSize       int

	ContentSize uint64
	DictID      uint32
}

type FrameDescError struct {
	reason string
}

func NewFrameDescError(reason string) error {
	return &FrameDescError{reason}
}

func (e FrameDescError) Error() string {
	return "Frame Descriptor: " + e.reason
}

const (
	lzLenFrameDescContentSize = 8
	lzLenFrameDescDictID      = 4
	lzLenFrameDescMax         = lzLenFrameDescContentSize + lzLenFrameDescDictID + 3
)

const (
	lzMaskVersion         byte = 0xc0
	lzMaskIndependent          = 0x20
	lzMaskBlockChecksum        = 0x10
	lzMaskContentSize          = 0x08
	lzMaskContentChecksum      = 0x04
	lzMaskReservedFLG          = 0x02
	lzMaskDictID               = 0x01
	lzMaskReservedBD           = 0x8f
	lzMaskBlockSize            = 0x70

	lzBlock64K  = 0x40
	lzBlock256K = 0x50
	lzBlock1M   = 0x60
	lzBlock4M   = 0x70
)

func ReadFrameDesc(r io.Reader) (f *FrameDesc, err error) {
	var b [lzLenFrameDescMax]byte
	var n int
	n, err = r.Read(b[:3])
	if err != nil {
		return
	}

	if b[0]&lzMaskVersion != 0x40 {
		err = NewFrameDescError("FLG: version number")
		return
	}
	if b[0]&lzMaskReservedFLG != 0 {
		err = NewFrameDescError("FLG: reserved bits")
		return
	}
	if b[1]&lzMaskReservedBD != 0 {
		err = NewFrameDescError("BD: reserved bits")
		return
	}

	f = &FrameDesc{
		Independent:        b[0]&lzMaskIndependent != 0,
		HasBlockChecksum:   b[0]&lzMaskBlockChecksum != 0,
		HasContentChecksum: b[0]&lzMaskContentChecksum != 0,
	}

	switch b[1] & lzMaskBlockSize {
	case lzBlock64K:
		f.BlockMaxSize = 64 << 10
	case lzBlock256K:
		f.BlockMaxSize = 256 << 10
	case lzBlock1M:
		f.BlockMaxSize = 1 << 20
	case lzBlock4M:
		f.BlockMaxSize = 4 << 20
	default:
		err = NewFrameDescError("BD: block max size")
		return
	}

	if b[0]&lzMaskContentSize != 0 {
		var optLen int
		optLen, err = r.Read(b[n : n+lzLenFrameDescContentSize])
		n += optLen
		if err != nil {
			return
		}
		f.ContentSize = leUint64(b[2:])
	}

	if b[0]&lzMaskDictID != 0 {
		var optLen int
		optLen, err = r.Read(b[n : n+lzLenFrameDescDictID])
		n += optLen
		if err != nil {
			return
		}
		f.DictID = leUint32(b[2:])
	}

	hChecksum := byte((xxhash.Checksum32(b[:n-1]) >> 8) & 0xff)
	if hChecksum != b[n-1] {
		err = NewFrameDescError("checksum")
		return
	}

	return f, nil
}

func leUint32(b []byte) uint32 {
	return uint32(b[0]) |
		uint32(b[1])<<8 |
		uint32(b[2])<<16 |
		uint32(b[3])<<24
}

func leUint64(b []byte) uint64 {
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<52
}
