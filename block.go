//
// block.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"fmt"
)

func DecompressBlock(in []byte, out []byte) (int, error) {
	var literals, matchLen, matchOfft, i, j, n int
	for i < len(in) {
		literals = int(in[i] >> 4)
		matchLen = int(in[i] & 0xf)
		i++

		if literals == 15 {
			for ; in[i] == 255 && i < len(in); i++ {
				if MaxInt-255 < literals {
					return -1, fmt.Errorf("DecompressBlock: literals length is too big")
				}
				literals += 255
			}
			literals += int(in[i])
			i++
		}
		//fmt.Fprintf(os.Stderr, "Seq literals: %d\n", literals)
		if literals != 0 {
			n = copy(out[j:], in[i:i+literals])
			if n != literals {
				return -1, fmt.Errorf("DecompressBlock: could not copy literals - small buffer")
			}
			i += literals
			j += literals
		}

		if i == len(in) {
			// EOF reached
			// parsing restrictions, check https://github.com/lz4/lz4/blob/dev/lib/lz4.c#L1171
			return j, nil
		}

		matchOfft = int(in[i]) | int(in[i+1])<<8
		if j < matchOfft {
			return -1, fmt.Errorf("DecompressBlock: malformed match offset")
		}
		i += 2

		if matchLen == 15 {
			for ; in[i] == 255 && i < len(in); i++ {
				if MaxInt-251 < matchLen {
					return -1, fmt.Errorf("DecompressBlock: match length is too big")
				}
				matchLen += 255
			}
			matchLen += int(in[i])
			i++
		}
		matchLen += 4 // additional 4 bytes - minmatch
		//fmt.Fprintf(os.Stderr, "Seq match: %d\n", matchLen)
		n = copy(out[j:], out[j-matchOfft:j-matchOfft+matchLen])
		if n != matchLen {
			return -1, fmt.Errorf("DecompressBlock: could not copy match - small buffer")
		}
		j += matchLen
	}
	return j, nil
}

type BlockDecoder struct {
	i, j       int
	in, out    []byte
	endOfBlock bool
}

func NewBlockDecoder(in, out []byte) *BlockDecoder {
	return &BlockDecoder{in: in, out: out}
}

func (d *BlockDecoder) Reset() {
	d.i = 0
	d.j = 0
	d.endOfBlock = false
}

func (d *BlockDecoder) DecompressBlock(blockLen int) (n int, err error) {
	for !d.endOfBlock {
		err = d.DecodeSequence(blockLen)
		if err != nil {
			return -1, err
		}
	}
	return d.j, nil
}

func (d *BlockDecoder) DecodeSequence(blockLen int) error {
	var n int
	// jinitial := d.j
	// log.Printf("token: %02x", d.in[d.i])
	lLen := int(d.in[d.i] >> 4)
	mLen := int(d.in[d.i]&0xf) + 4
	d.i++
	if lLen == 15 {
		for ; d.in[d.i] == 255 && d.i < len(d.in); d.i++ {
			// log.Printf("lLen : %02x", d.in[d.i])
			lLen += 255
		}
		lLen += int(d.in[d.i])
		// log.Printf("lLen : %02x", d.in[d.i])
		if lz4MB < lLen { // could not be greater then max block len
			return fmt.Errorf("DecompressSequence: malformed input")
		}
		d.i++
	}
	// log.Printf("copy %d literals", lLen)
	if lLen != 0 {
		n = copy(d.out[d.j:], d.in[d.i:d.i+lLen])
		if n != lLen {
			return fmt.Errorf("DecompressSequence: buffer is too small")
		}
		d.i += lLen
		d.j += lLen
	}

	// log.Printf("current pos %d/%d", d.i, blockLen)
	if d.i == blockLen { // EOF reached
		d.endOfBlock = true
		return nil
	}
	// log.Printf("mOfft: %02x%02x", d.in[d.i], d.in[d.i+1])
	mOfft := int(d.in[d.i]) | int(d.in[d.i+1])<<8
	// log.Printf("match offset: %d", mOfft)
	if d.j < mOfft {
		return fmt.Errorf("DecompressSequence: malformed input")
	}
	d.i += 2
	if mLen == 19 {
		for ; d.in[d.i] == 255 && d.i < len(d.in); d.i++ {
			// log.Printf("mLen : %02x", d.in[d.i])
			mLen += 255
		}
		// log.Printf("mLen : %02x", d.in[d.i])
		mLen += int(d.in[d.i])
		if lz4MB < mLen { // could not be greater then max block len
			return fmt.Errorf("DecompressSequence: malformed input")
		}
		d.i++
	}
	for mOfft < mLen {
		// log.Printf("match copy %d literals at offt %d partial", mOfft, mOfft)
		n = copy(d.out[d.j:], d.out[d.j-mOfft:d.j])
		if n != mOfft {
			return fmt.Errorf("DecompressSequence: buffer is too small")
		}
		mLen -= mOfft
		d.j += n
	}
	// log.Printf("match copy %d literals at offt %d", mLen, mOfft)
	n = copy(d.out[d.j:], d.out[d.j-mOfft:d.j-mOfft+mLen])
	if n != mLen {
		return fmt.Errorf("DecompressSequence: buffer is too small")
	}
	d.j += mLen
	// fmt.Fprintln(os.Stderr, hex.Dump(d.out[jinitial:d.j]))
	return nil
}
