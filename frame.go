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

type Frame struct {
	*FrameDesc
	hash        *xxhash.XXHash32
	in, out     []byte
	bContentLen int
}

func NewFrame(r io.Reader) (f *Frame, err error) {
	f = new(Frame)
	f.FrameDesc, err = ReadFrameDesc(r)
	if err != nil {
		return nil, err
	}
	f.in = make([]byte, f.BlockMaxSize)
	f.out = make([]byte, f.BlockMaxSize)
	if f.HasContentChecksum {
		f.hash = xxhash.New32()
	}
	return f, nil
}

func (f *Frame) DecompressBlock(r io.Reader) ([]byte, error) {
	ulen, err := readUint32(r, f.out[:4])
	if err != nil {
		return nil, err
	}

	compressed := true
	if (ulen & 0x80000000) != 0 {
		ulen &= 0x7fffffff
		compressed = false
	}

	bLen := int(ulen)
	if bLen > len(f.in) {
		return nil, fmt.Errorf("invalid block size")
	}

	fmt.Fprintf(os.Stderr,
		"Read Block<Len: %d Compressed: %v BlockChecksum: %v>\n",
		bLen, compressed, blockChecksum,
	)

	if bLen == 0 { // EndMark
		return nil, nil
	}

	err = read(r, f.in[:bLen])
	if err != nil {
		return nil, err
	}

	if f.HasBlockChecksum {
		bChecksum, err := readUint32(r, f.out[:4])
		if err != nil {
			return nil, err
		}
		must := xxhash.Checksum32(f.in[:bLen])
		if bChecksum != must {
			return nil, fmt.Errorf("block checksum mismatch")
		}
	}

	if compressed {
		f.bContentLen, err = DecompressBlock2(f.in[:bLen], f.out)
		if err != nil {
			return nil, err
		}
	} else {
		f.bContentLen = copy(f.out, f.in[:bLen])
	}

	if f.hash != nil {
		_, err = f.hash.Write(f.out[:f.bContentLen])
		if err != nil {
			return nil, err
		}
	}

	return nil
}

type Reader struct {
	r io.Reader

	pos    int
	frame  *Frame
	intBuf [8]byte
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if r.frame == nil {
		// read lz4 frame
		magic, err := readUint32(r.r, r.intBuf[:])
		if err != nil {
			return err
		}
		if magic != lzMagic {
			return 4, fmt.Errorf("bad frame magic")
		}
		r.frame, err = NewFrame(r.r)
		if err != nil {
			return
		}
	}
	var n int
	for n != len(b) {
		if r.pos <= r.frame.bContentLen {
			err = r.frame.DecompressBlock(r.r)
			if err != nil {
				return
			}
			r.pos = 0
		}
		n += copy(b[n:], r.frame.out[r.pos:])
	}

}

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
			n, err = DecompressBlock2(in[:bLen], out)
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
