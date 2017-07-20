//
// frame.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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

func (d *Decompressor) Decompress(r io.Reader, w io.Writer) (err error) {
	d.r = r
	var magic uint32
	magic, err = readUint32(d.r, d.buf[:])
	if magic != lzMagic {
		err = fmt.Errorf("Decompress: Frame magic not match")
		return
	}

	var desc *FrameDesc
	desc, err = ReadFrameDesc(d.r)
	if err != nil {
		return
	}

	jb, _ := json.MarshalIndent(desc, "", "  ")
	fmt.Fprintln(os.Stderr, string(jb))

	in := make([]byte, desc.BlockMaxSize)
	out := make([]byte, desc.BlockMaxSize)

	var cMust *xxhash.XXHash32
	if desc.HasContentChecksum {
		cMust = xxhash.New32()
	}

	var bLen, n int
	var compressed bool
	decoder := NewBlockDecoder(in, out)
	for {
		bLen, compressed, err = d.readBlockLen(desc.BlockMaxSize)
		if err != nil {
			return
		}
		fmt.Fprintf(os.Stderr, "Block Len: %d, compressed: %v\n", bLen, compressed)
		if bLen == 0 { // EndMark
			break
		}
		err = read(d.r, in[:bLen])
		if err != nil {
			return
		}
		if desc.HasBlockChecksum {
			var bChecksum uint32
			bChecksum, err = readUint32(d.r, d.buf[:])
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
			n, err = decoder.DecompressBlock(bLen)
			if err != nil {
				return
			}
		} else {
			n = copy(out, in[:bLen])
		}
		if desc.HasContentChecksum {
			fmt.Fprintf(os.Stderr, "count: %d\n%s\n", n, out[n-50:n])
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
		cChecksum, err = readUint32(d.r, d.buf[:])
		if err != nil {
			return err
		}
		if cChecksum != cMust.Sum32() {
			err = fmt.Errorf(
				"DecodeFrame: Content checksum mismatch %04x != %04x",
				cMust.Sum32(), cChecksum)
			return
		}
	}

	return nil
}

func (d *Decompressor) readBlockLen(maxLen int) (len int, compressed bool, err error) {
	ulen, err := readUint32(d.r, d.buf[:])
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
