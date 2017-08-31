//
// desc_test.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFrameHeaderErrors(t *testing.T) {
	var tt = []struct {
		name string
		b    []byte
	}{
		{
			"not enough input",
			[]byte{},
		},
		{
			"bad version",
			[]byte{0x00, 0x00, 0x00},
		},
		{
			"bad reserved bits",
			[]byte{0x43, 0x8f, 0x00},
		},
		{
			"bad block max size",
			[]byte{0x40, 0x30, 0xc0},
		},
		{
			"not enought input for content size",
			[]byte{0x48, 0x40, 0xfd},
		},
		{
			"checksum mismatch",
			[]byte{0x40, 0x40, 0xff},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.b)
			_, err := ReadFrameDesc(r)
			if err == nil {
				t.Fatalf("Expecected an error, got nil!")
			}
		})
	}
}

func TestReadFrameHeader(t *testing.T) {
	var tt = []struct {
		name  string
		b     []byte
		eDesc FrameDesc
	}{
		{
			"all false 64K",
			[]byte{0x40, 0x40, 0xc0},
			FrameDesc{
				Independent:        false,
				HasBlockChecksum:   false,
				HasContentChecksum: false,
				BlockMaxSize:       64 << 10,
				ContentSize:        0,
			},
		},
		{
			"all false 256K",
			[]byte{0x40, 0x50, 0x77},
			FrameDesc{
				Independent:        false,
				HasBlockChecksum:   false,
				HasContentChecksum: false,
				BlockMaxSize:       256 << 10,
				ContentSize:        0,
			},
		},
		{
			"all true 1M",
			[]byte{0x74, 0x60, 0xd9},
			FrameDesc{
				Independent:        true,
				HasBlockChecksum:   true,
				HasContentChecksum: true,
				BlockMaxSize:       1 << 20,
				ContentSize:        0,
			},
		},
		{
			"all true 4M + content size",
			[]byte{0x7c, 0x70, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfd},
			FrameDesc{
				Independent:        true,
				HasBlockChecksum:   true,
				HasContentChecksum: true,
				BlockMaxSize:       4 << 20,
				ContentSize:        10,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.b)
			f, err := ReadFrameDesc(r)
			if err != nil {
				t.Fatal(err)
			}
			if f.Independent != tc.eDesc.Independent ||
				f.HasBlockChecksum != tc.eDesc.HasBlockChecksum ||
				f.HasContentChecksum != tc.eDesc.HasContentChecksum ||
				f.BlockMaxSize != tc.eDesc.BlockMaxSize ||
				f.ContentSize != tc.eDesc.ContentSize {
				b1, _ := json.MarshalIndent(&tc.eDesc, "", "  ")
				b2, _ := json.MarshalIndent(f, "", "  ")
				t.Errorf(
					"Mismatch\nExpected:\n%s\nHas:\n%s",
					string(b1), string(b2),
				)
			}
		})
	}
}
