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

const lzMagic uint32 = 0x184d2204

type Decompressor struct {
	r io.Reader

	buf [8]byte
}

func NewDecompressor() *Decompressor {
	return &Decompressor{}
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
	if magic != lzMagic {
		err = fmt.Errorf("Decompress: Frame magic not match")
		return
	}

	var desc *FrameDesc
	desc, err = ReadFrameDesc(d.r)
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
