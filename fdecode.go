//
// fdecoder.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"bytes"
	"fmt"
	"io"

	"github.com/OneOfOne/xxhash"
)

type BlockDecoder struct {
	r     io.Reader
	frame *FrameDesc
	buf   *BlockBuffer
}

func NewBlockDecoder(r io.Reader) (*BlockDecoder, error) {
	var b [4]byte
	_, err := r.Read(b[:])
	if err != nil {
		return nil, err
	}

	if bytes.Compare([]byte{0x04, 0x22, 0x4d, 0x18}, b[:]) != 0 {
		return nil, fmt.Errorf("bad magic")
	}

	d := &BlockDecoder{r: r}
	d.frame, err = ReadFrameDesc(r)
	if err != nil {
		return nil, err
	}

	d.buf = NewBlockBuffer(d.frame.BlockMaxSize)
	return d, nil
}

func (d *BlockDecoder) BlockMaxSize() int {
	return d.frame.BlockMaxSize
}

func (d *BlockDecoder) DecodeBlock(b []byte) (n int, err error) {
	if d.buf.Empty() {
		err = d.readBlock()
		if err != nil {
			return
		}
	}

	if d.buf.Compressed {
		n, err = DecompressBlock(d.buf.Bytes(), b)
	} else {
		n = copy(b, d.buf.Bytes())
		if n < len(b) {
			err = fmt.Errorf("buffer is too small")
		}
	}
	if err != nil {
		return
	}

	d.buf.Reset()
	return n, nil
}

func (d *BlockDecoder) SkipBlock() error {
	err := d.readBlock()
	if err != nil {
		return err
	}
	return nil
}

type BlockBuffer struct {
	Buf        []byte
	Len        int
	Compressed bool
}

func NewBlockBuffer(size int) *BlockBuffer {
	return &BlockBuffer{
		Buf: make([]byte, size),
	}
}

func (b *BlockBuffer) Fill(r io.Reader, len uint32) (err error) {
	b.Len, err = r.Read(b.Buf[:len])
	if err != nil {
		b.Len = 0
		return
	}
	return nil
}

func (b *BlockBuffer) Bytes() []byte {
	return b.Buf[:b.Len]
}

func (b *BlockBuffer) Reset() {
	b.Len = 0
	b.Compressed = false
}

func (b *BlockBuffer) Checksum32() uint32 {
	return xxhash.Checksum32(b.Buf[:b.Len])
}

func (b *BlockBuffer) Empty() bool {
	return b.Len == 0
}

func (d *BlockDecoder) readBlock() (err error) {
	u, err := readUint32(d.r)
	if err != nil {
		return err
	}

	d.buf.Compressed = u&0x80000000 == 0
	err = d.buf.Fill(d.r, u&0x7fffffff)
	if err != nil {
		return err
	}

	if d.frame.HasBlockChecksum {
		h, err := readUint32(d.r)
		if err != nil {
			return err
		}
		if h != d.buf.Checksum32() {
			return fmt.Errorf("block checksum mismatch")
		}
	}

	return nil
}

func readUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	_, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	return leUint32(b[:]), nil
}
