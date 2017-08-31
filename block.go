//
// block.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

type InvalidInput struct {
	desc string
}

func (e InvalidInput) Error() string {
	return "Invalid input: " + e.desc
}

// DecompressBlock decompress all sequences in compressed lz4 block
// return count of decompressed bytes in out slice
func DecompressBlock(in, out []byte) (j int, err error) {
	var i, n, lLen, mLen, mOfft int
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
				return j, &InvalidInput{"malformed input: literals len"}
			}
			if n = copy(out[j:], in[i:i+lLen]); n != lLen {
				return j, &InvalidInput{"could not copy literals"}
			}
			i += lLen
			j += lLen
		}
		if i == len(in) { // reached end of block
			return j, nil
		}
		if i+1 == len(in) {
			return j, &InvalidInput{"malformed input: match offset"}
		}
		mOfft = int(in[i]) | int(in[i+1])<<8
		if i += 2; j < mOfft || mOfft == 0 {
			return j, &InvalidInput{"malformed input: match offset"}
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
				return j, &InvalidInput{"could not copy match"}
			}
			j += mOfft
		}
		if n = copy(out[j:], out[j-mOfft:j-mOfft+mLen]); n != mLen {
			return j, &InvalidInput{"could not copy match"}
		}
		j += mLen
	}
	return j, nil
}
