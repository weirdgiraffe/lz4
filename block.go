//
// block.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

const AllSequences = -1

type InvalidInput struct {
	desc string
}

func (e InvalidInput) Error() string {
	return "Invalid input: " + e.desc
}

// DecompressBlock decompress at least seqCount sequences from lz4 compressed block
// return offsets in input and output buffers after last successfuly decompressed sequence
func DecompressBlock(in, out []byte, seqCount int) (i, j int, err error) {
	var n, lLen, mLen, mOfft, seq int
	for i < len(in) && (seqCount < 0 || seq < seqCount) {
		istart := i
		jstart := j
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
				return istart, jstart, &InvalidInput{"malformed input: literals len"}
			}
			if n = copy(out[j:], in[i:i+lLen]); n != lLen {
				return istart, jstart, &InvalidInput{"could not copy literals"}
			}
			i += lLen
			j += lLen
		}
		if i == len(in) { // reached end of block
			return i, j, nil
		}
		if i+1 == len(in) {
			return istart, jstart, &InvalidInput{"malformed input: match offset"}
		}
		mOfft = int(in[i]) | int(in[i+1])<<8
		if i += 2; j < mOfft || mOfft == 0 {
			return istart, jstart, &InvalidInput{"malformed input: match offset"}
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
				return istart, jstart, &InvalidInput{"could not copy match"}
			}
			j += mOfft
		}
		if n = copy(out[j:], out[j-mOfft:j-mOfft+mLen]); n != mLen {
			return istart, jstart, &InvalidInput{"could not copy match"}
		}
		j += mLen
		seq++
	}
	return i, j, nil
}
