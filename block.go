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
		n = copy(out[j:], out[j-matchOfft:j-matchOfft+matchLen])
		if n != matchLen {
			return -1, fmt.Errorf("DecompressBlock: could not copy match - small buffer")
		}
		j += matchLen
	}
	return j, nil
}

type InvalidInput struct {
	desc string
}

func (e InvalidInput) Error() string {
	return "Invalid input: " + e.desc
}

func DecompressBlock2(in, out []byte) (n int, err error) {
	var i, j, lLen, mLen, mOfft int
	for i < len(in) {
		lLen = int(in[i] >> 4)
		mLen = int(in[i]&0xf) + 4
		if i++; lLen != 0 {
			if lLen == 15 {
				for i < len(in) {
					lLen += int(in[i])
					if i++; in[i-1] != 255 {
						break
					}
				}
				if len(in) < lLen { // could not be greater then block len
					return -1, &InvalidInput{"literals len > block len"}
				}
			}
			if n = copy(out[j:], in[i:i+lLen]); n != lLen {
				return -1, &InvalidInput{"insufficient buffer"}
			}
			i += lLen
			j += lLen
		}
		if i == len(in) { // EOF reached
			return j, nil
		}
		mOfft = int(in[i]) | int(in[i+1])<<8
		if i += 2; j < mOfft {
			return -1, &InvalidInput{"match offset otside buffer"}
		}
		if mLen == 19 {
			for i < len(in) {
				mLen += int(in[i])
				if i++; in[i-1] != 255 {
					break
				}
			}
			if lz4MB < lLen { // could not be greater then max block len
				return -1, &InvalidInput{"match len > block len"}
			}
		}
		for ; mOfft < mLen; mLen -= mOfft {
			if n = copy(out[j:], out[j-mOfft:j]); n != mOfft {
				return -1, &InvalidInput{"insufficient buffer"}
			}
			j += mOfft
		}
		if n = copy(out[j:], out[j-mOfft:j-mOfft+mLen]); n != mLen {
			return -1, &InvalidInput{"insufficient buffer"}
		}
		j += mLen
	}
	return j, nil
}
