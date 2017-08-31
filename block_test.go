//
// block_test.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDecompressBlockOK(t *testing.T) {
	tt := []struct {
		name     string
		seq      []byte
		expected []byte
	}{
		{
			name:     "minimal last sequence",
			seq:      []byte{0x10, 0x01},
			expected: []byte{0x01},
		},
		{
			name:     "overlapping match",
			seq:      []byte{0x42, 0x01, 0x02, 0x03, 0x04, 0x04, 0x00},
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01, 0x02},
		},
		{
			name:     "normal match",
			seq:      []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x04, 0x00},
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x01, 0x02, 0x03, 0x04},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make([]byte, len(tc.expected))
			n, err := DecompressBlock(tc.seq, out)
			if err != nil {
				t.Fatalf("Failed to decompress block: %v", err)
			}
			if 0 != bytes.Compare(out[:n], tc.expected) {
				t.Errorf(
					"Decompressed block mismatch\n      got: %s\n expected: %s",
					hex.EncodeToString(out[:n]),
					hex.EncodeToString(tc.expected),
				)
			}
		})
	}
}

func TestDecompressBlockErr(t *testing.T) {
	tt := []struct {
		name   string
		seq    []byte
		outLen int
	}{
		{
			name:   "not enough bytes for literals len",
			seq:    []byte{0xf0},
			outLen: 10,
		},
		{
			name:   "not enough literals",
			seq:    []byte{0x40, 0x01},
			outLen: 10,
		},
		{
			name:   "not enough bytes for match offset",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x00},
			outLen: 10,
		},
		{
			name:   "output buffer is too smal for literals",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04},
			outLen: 3,
		},
		{
			name:   "match offset outside out buffer",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x08, 0x00},
			outLen: 10,
		},
		{
			name:   "null match offset",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x00, 0x00},
			outLen: 7,
		},
		{
			name:   "not enough bytes for match len",
			seq:    []byte{0x4f, 0x01, 0x02, 0x03, 0x04, 0x04, 0x00, 0xaa},
			outLen: 10,
		},
		{
			name:   "out buffer is too small for overlapping match",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x02, 0x00},
			outLen: 7,
		},
		{
			name:   "out buffer is too small for match",
			seq:    []byte{0x40, 0x01, 0x02, 0x03, 0x04, 0x04, 0x00},
			outLen: 7,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := make([]byte, tc.outLen)
			_, err := DecompressBlock(tc.seq, out)
			if err == nil {
				t.Fatalf("Expecected an error, got nil!")
			}
		})
	}
}
