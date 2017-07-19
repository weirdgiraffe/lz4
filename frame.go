//
// frame.go
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

const lz4Magic uint32 = 0x184d2204

const (
	lz64KB  = 65536
	lz256KB = 262144
	lz1MB   = 1048576
	lz4MB   = 4194304
)

type Decompressor struct {
	r io.Reader

	buf [8]byte
}

func NewDecompressor() *Decompressor {
	return &Decompressor{}
}

type FrameDesc struct {
	Independent        bool
	HasBlockChecksum   bool
	HasContentChecksum bool
	BlockMaxSize       int
	ContentSize        uint64
}

const (
	lzFLGByte         = 0
	lzBDByte          = 1
	lzMaxFrameDescLen = 11
)

func readFrameDesc(r io.Reader) (f *FrameDesc, err error) {
	var b [lzMaxFrameDescLen]byte
	// read FLG byte + BD byte + (HC byte or first byte of ContentSize)
	err = read(r, b[:3])
	if err != nil {
		return
	}
	// check version is 01
	if b[lzFLGByte]&0xc0 != 0x40 {
		err = fmt.Errorf("FrameDesc: version must be 01")
		return
	}
	// check reserved bits are 0
	if b[lzFLGByte]&0x03 != 0 && b[lzBDByte]&0x8f != 0 {
		err = fmt.Errorf("FrameDesc: reserved bits must be zero")
		return
	}
	bSize := lzBlockMaxSize(b[lzBDByte] & 0x70 >> 4)
	if bSize == -1 {
		err = fmt.Errorf("FrameDesc: unsupported Block Maximum size")
		return
	}
	// optional ContentSize field
	var contentSize uint64
	i := lzBDByte + 1
	if b[lzFLGByte]&0x08 != 0 {
		// first ContentSize byte is already in b, read rest + HC byte
		err = read(r, b[i+1:i+9])
		if err != nil {
			return
		}
		contentSize = leUint64(b[i : i+8])
		i += 8
	}
	// HC byte
	hChecksum := byte((xxhash.Checksum32(b[:i]) >> 8) & 0xff)
	if hChecksum != b[i] {
		err = fmt.Errorf(
			"FrameDesc: checksum mismatch %02x != %02x",
			hChecksum, b[i],
		)
		return
	}
	return &FrameDesc{
		Independent:        b[lzFLGByte]&0x20 != 0,
		HasBlockChecksum:   b[lzFLGByte]&0x10 != 0,
		HasContentChecksum: b[lzFLGByte]&0x04 != 0,
		BlockMaxSize:       bSize,
		ContentSize:        contentSize,
	}, nil
}

func (d *Decompressor) readBlock(block []byte) (err error) {
	err = read(d.r, block)
	if err != nil {
		return
	}
	return nil
}

func (d *Decompressor) Decompress(r io.Reader, w io.Writer) (err error) {
	d.r = r
	var magic uint32
	magic, err = d.readUint32()
	if magic != lz4Magic {
		err = fmt.Errorf("Decompress: Frame magic not match")
		return
	}

	var desc *FrameDesc
	desc, err = readFrameDesc(d.r)
	if err != nil {
		return
	}

	in := make([]byte, desc.BlockMaxSize)
	out := make([]byte, desc.BlockMaxSize)

	var cMust *xxhash.XXHash32
	if desc.HasContentChecksum {
		cMust = xxhash.New32()
	}

	var bLen, n int
	var compressed bool
	for {
		bLen, compressed, err = d.readBlockLen(desc.BlockMaxSize)
		if err != nil {
			return
		}
		if bLen == 0 { // EndMark
			break
		}
		err = d.readBlock(in[:bLen])
		if err != nil {
			return
		}

		if desc.HasBlockChecksum {
			var bChecksum uint32
			bChecksum, err = d.readUint32()
			if err != nil {
				return
			}
			must := xxhash.Checksum32(in[:bLen])
			if bChecksum != must {
				err = fmt.Errorf("DecodeFrame: Block checksum mismatch")
				return
			}
		}

		if compressed {
			n, err = DecompressBlock(in[:bLen], out)
			if err != nil {
				return
			}
		} else {
			n = copy(out, in[:bLen])
		}

		if desc.HasContentChecksum {
			_, err = cMust.Write(out[:n])
			if err != nil {
				return err
			}
		}

		err = write(w, out[:n])
		if err != nil {
			return
		}
	}

	if desc.HasContentChecksum {
		var cChecksum uint32
		cChecksum, err = d.readUint32()
		if err != nil {
			return err
		}
		if cChecksum != cMust.Sum32() {
			err = fmt.Errorf("DecodeFrame: Content checksum mismatch")
			return
		}
	}

	return nil
}

func (d *Decompressor) readUint64() (uint64, error) {
	err := read(d.r, d.buf[:8])
	if err != nil {
		return 0, err
	}
	return leUint64(d.buf[:8]), nil
}

func (d *Decompressor) readUint32() (uint32, error) {
	err := read(d.r, d.buf[:4])
	if err != nil {
		return 0, err
	}
	return leUint32(d.buf[:4]), nil
}

func read(r io.Reader, b []byte) error {
	n, err := r.Read(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Could not read enough data (need %d) got EOF", len(b))
	}
	return err
}

func write(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Decomressor: schrinked write")
	}
	return err
}

func (d *Decompressor) readBlockLen(maxLen int) (len int, compressed bool, err error) {
	ulen, err := d.readUint32()
	if err != nil {
		return
	}
	compressed = (ulen & 0x80000000) == 0
	len = int(ulen & 0x7fffffff)
	if len > maxLen {
		err = fmt.Errorf("DecompressFrame: malformed block size")
		return
	}
	return len, compressed, nil
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
