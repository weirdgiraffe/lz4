//
// common.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

const MaxInt = int(^uint(0) >> 1)

func leUint32(b []byte) uint32 {
	return uint32(b[0]) |
		uint32(b[1])<<8 |
		uint32(b[2])<<16 |
		uint32(b[3])<<24
}

func leUint64(b []byte) uint64 {
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<52
}
