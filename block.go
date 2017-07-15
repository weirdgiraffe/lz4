//
// block.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import "log"

func Decompress(in []byte, out []byte) int {
	token := in[0]
	i := 1 // token took position 0

	literals := seqLen(token>>4, in[i:])
	i += seqLenAdditionalConsumed(literals)
	log.Printf("literals: %d", literals)

	copy(out, in[i:i+literals])
	i += literals

	offt := int(in[i]) | int(in[i+1])<<8
	i += 2
	log.Printf("offset: %d", offt)

	match := seqLen(token&0xf, in[i:])
	i += seqLenAdditionalConsumed(match)
	match += 4 // minmatch is 4
	log.Printf("match: %d", match)

	copy(out[literals:], out[literals-offt:literals-offt+match])
	return literals + match
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
