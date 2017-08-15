//
// block.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

const maxBlockSize = 4 << 10

type InvalidInput struct {
	desc string
}

func (e InvalidInput) Error() string {
	return "Invalid input: " + e.desc
}

func DecompressBlock(in, out []byte) (n int, err error) {
	var i, j, lLen, mLen, mOfft int
	for i < len(in) {
		lLen = int(in[i] >> 4)
		mLen = int(in[i]&0xf) + 4
		i++
		if lLen != 0 {
			if lLen == 15 {
				for i < len(in) {
					lLen += int(in[i])
					if i++; in[i-1] != 255 {
						break
					}
				}
			}
			if i == len(in) || i+lLen > len(in) {
				return -1, &InvalidInput{"malformed input: literals len"}
			}
			if n = copy(out[j:], in[i:i+lLen]); n != lLen {
				return -1, &InvalidInput{"could not copy literals"}
			}
			i += lLen
			j += lLen
		}

		if i == len(in) { // reached end of block
			return j, nil
		}
		if i+1 == len(in) {
			return -1, &InvalidInput{"malformed input: match offset"}
		}
		mOfft = int(in[i]) | int(in[i+1])<<8
		if i += 2; j < mOfft {
			return -1, &InvalidInput{"malformed input: match offset"}
		}
		if mLen == 19 {
			for i < len(in) {
				mLen += int(in[i])
				if i++; in[i-1] != 255 {
					break
				}
			}
		}
		for ; mOfft < mLen; mLen -= mOfft {
			if n = copy(out[j:], out[j-mOfft:j]); n != mOfft {
				return -1, &InvalidInput{"could not copy match"}
			}
			j += mOfft
		}
		if n = copy(out[j:], out[j-mOfft:j-mOfft+mLen]); n != mLen {
			return -1, &InvalidInput{"could not copy match"}
		}
		j += mLen
	}
	return j, nil
}
