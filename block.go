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

func Decompress(in []byte, out []byte) (int, error) {
	token := in[0]
	i := 1 // token has taken position 0

	literals := seqLen(token>>4, in[i:])
	if len(out) < literals {
		return 0, fmt.Errorf("output buffer is to small")
	}
	i += seqLenAdditionalConsumed(literals)

	i += copy(out, in[i:i+literals])

	offt := int(in[i]) | int(in[i+1])<<8
	if literals-offt < 0 {
		return i, fmt.Errorf("offset must be inside literals")
	}
	i += 2

	match := seqLen(token&0xf, in[i:]) + 4 // additional 4 bytes - minmatch
	if len(out) < literals+match {
		return 0, fmt.Errorf("output buffer is to small")
	}
	i += seqLenAdditionalConsumed(match - 4)

	return literals + copy(out[literals:], out[literals-offt:literals-offt+match]), nil
}

func seqLenAdditionalConsumed(l int) int {
	if l < 15 {
		return 0
	}
	return (l / 255) + 1
}

func seqLen(initialLen byte, b []byte) (l int) {
	l = int(initialLen)
	if l == 0 {
		return 1
	}
	if l != 15 {
		return
	}
	for i := range b {
		l += int(b[i])
		if b[i] != 255 {
			return
		}
	}
	panic("wrong length")
}
