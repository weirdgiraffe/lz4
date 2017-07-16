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
	i, j := 0, 0
	for i < len(in) {
		ni, nj, err := DecompressSequence(in[i:], out, j)
		if err != nil {
			return -1, err
		}
		i += ni
		j += nj
	}
	return j, nil
}

func DecompressSequence(in []byte, out []byte, offt int) (int, int, error) {
	token := in[0]

	literals, pos := varlen(token>>4, in[1:])
	pos += 1 // additional 1 byte for token itself

	n := copy(out[offt:], in[pos:pos+literals])
	if n != literals {
		return -1, -1, fmt.Errorf("literals: in[] or out[] is too small")
	}
	pos += literals

	if pos == len(in) {
		// EOF reached
		// parsing restrictions, check https://github.com/lz4/lz4/blob/dev/lib/lz4.c#L1171
		return pos, literals, nil
	}

	matchOfft := int(in[pos]) | int(in[pos+1])<<8
	if (offt+literals)-matchOfft < 0 {
		return -1, -1, fmt.Errorf("match offset must be inside out[]")
	}
	pos += 2

	match, n := varlen(token&0xf, in[pos:])
	match += 4 // additional 4 bytes - minmatch
	pos += n

	n = copy(out[offt+literals:], out[offt+literals-matchOfft:offt+literals-matchOfft+match])
	if n != match {
		return -1, -1, fmt.Errorf("match: in[] or out[] is too small")
	}
	return pos, literals + match, nil
}

func varlen(initialLen byte, b []byte) (l, n int) {
	l = int(initialLen)
	if l == 0 {
		return
	}
	if l != 15 {
		return
	}
	for i := range b {
		l += int(b[i])
		n++
		if b[i] != 255 {
			return
		}
	}
	panic("wrong length")
}
