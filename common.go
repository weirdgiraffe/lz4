//
// common.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"fmt"
	"io"
)

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

func read(r io.Reader, b []byte) error {
	n, err := r.Read(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Could not read enough data (need %d) got EOF", len(b))
	}
	return err
}

func write(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("Decomressor: schrinked write")
	}
	return err
}

func readUint64(r io.Reader, buf []byte) (uint64, error) {
	err := read(r, buf[:8])
	if err != nil {
		return 0, err
	}
	return leUint64(buf[:8]), nil
}

func readUint32(r io.Reader, buf []byte) (uint32, error) {
	err := read(r, buf[:4])
	if err != nil {
		return 0, err
	}
	return leUint32(buf[:4]), nil
}
