//
// frame_test.go
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

func TestReadFrameHeader(t *testing.T) {
	var tc = []struct {
		b     []byte
		eErr  bool
		eDesc FrameDesc
	}{
		{ // 1
			[]byte{0x40, 0x40, 0xc0},
			false,
			FrameDesc{
				Independent:        false,
				HasBlockChecksum:   false,
				HasContentChecksum: false,
				BlockMaxSize:       lz64KB,
				ContentSize:        0,
			},
		},
		{ // 2
			[]byte{0x40, 0x50, 0x77},
			false,
			FrameDesc{
				Independent:        false,
				HasBlockChecksum:   false,
				HasContentChecksum: false,
				BlockMaxSize:       lz256KB,
				ContentSize:        0,
			},
		},
		{ // 3
			[]byte{0x74, 0x60, 0xd9},
			false,
			FrameDesc{
				Independent:        true,
				HasBlockChecksum:   true,
				HasContentChecksum: true,
				BlockMaxSize:       lz1MB,
				ContentSize:        0,
			},
		},
		{ // 4
			[]byte{0x7c, 0x70, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfd},
			false,
			FrameDesc{
				Independent:        true,
				HasBlockChecksum:   true,
				HasContentChecksum: true,
				BlockMaxSize:       lz4MB,
				ContentSize:        10,
			},
		},
		{ // 5
			[]byte{}, // not enough input
			true,
			FrameDesc{},
		},
		{ // 6
			[]byte{0x00, 0x00, 0x00}, // bad version
			true,
			FrameDesc{},
		},
		{ // 7
			[]byte{0x43, 0x8f, 0x00}, // reserved bits
			true,
			FrameDesc{},
		},
		{ // 8
			[]byte{0x40, 0x30, 0xc0}, // bad block size
			true,
			FrameDesc{},
		},
		{ // 9
			[]byte{0x48, 0x40, 0xfd}, // not enought input for content size
			true,
			FrameDesc{},
		},
		{ // 10
			[]byte{0x40, 0x40, 0xff}, // desc checksum mismatch
			true,
			FrameDesc{},
		},
	}
	for i := range tc {
		r := bytes.NewReader(tc[i].b)
		f, err := readFrameDesc(r)
		if err != nil {
			if !tc[i].eErr {
				t.Fatalf(
					"tc #%d Unexpected error: %v",
					i+1, err,
				)
			}
			continue
		}
		if f.Independent != tc[i].eDesc.Independent ||
			f.HasBlockChecksum != tc[i].eDesc.HasBlockChecksum ||
			f.HasContentChecksum != tc[i].eDesc.HasContentChecksum ||
			f.BlockMaxSize != tc[i].eDesc.BlockMaxSize ||
			f.ContentSize != tc[i].eDesc.ContentSize {
			b1, _ := json.MarshalIndent(&tc[i].eDesc, "", "  ")
			b2, _ := json.MarshalIndent(f, "", "  ")
			t.Errorf(
				"tc #%d Unexpected Frame Discriptor:\nExpected:\n%s\nHas:\n%s",
				i+1, string(b1), string(b2),
			)
		}
	}

}
