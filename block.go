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

const MaxInt = int(^uint(0) >> 1)

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
