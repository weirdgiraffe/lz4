//
// block_test.go
// Copyright (C) 2017 weirdgiraffe <giraffe@cyberzoo.xyz>
//
// Distributed under terms of the MIT license.
//

package lz4

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var testLoremLZ4 = []byte{
	0x04, 0x22, 0x4d, 0x18, // magic
	0x64,                   // FLG: 01 - version 1 - independent 0 - block checksum 0-content size 1 - content checksum 00 - reserved
	0x40,                   //  BD: 0 100 - block size 64K 0000
	0xa7,                   //  HC: header checksum
	0xa5, 0x01, 0x00, 0x00, // block is compressed and block size  421
	0xf2, 0x57, 0x4c, 0x6f, 0x72, 0x65, 0x6d, 0x20, 0x69, 0x70, 0x73, 0x75, 0x6d, 0x20, 0x64, 0x6f,
	0x6c, 0x6f, 0x72, 0x20, 0x73, 0x69, 0x74, 0x20, 0x61, 0x6d, 0x65, 0x74, 0x2c, 0x20, 0x63, 0x6f,
	0x6e, 0x73, 0x65, 0x63, 0x74, 0x65, 0x74, 0x75, 0x72, 0x20, 0x61, 0x64, 0x69, 0x70, 0x69, 0x73,
	0x63, 0x69, 0x6e, 0x67, 0x20, 0x65, 0x6c, 0x69, 0x74, 0x2c, 0x20, 0x73, 0x65, 0x64, 0x20, 0x64,
	0x6f, 0x20, 0x65, 0x69, 0x75, 0x73, 0x6d, 0x6f, 0x64, 0x20, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72,
	0x20, 0x69, 0x6e, 0x63, 0x69, 0x64, 0x69, 0x64, 0x75, 0x6e, 0x74, 0x20, 0x75, 0x74, 0x20, 0x6c,
	0x61, 0x62, 0x6f, 0x72, 0x65, 0x20, 0x65, 0x74, 0x5b, 0x00, 0xf0, 0x0e, 0x65, 0x20, 0x6d, 0x61,
	0x67, 0x6e, 0x61, 0x20, 0x61, 0x6c, 0x69, 0x71, 0x75, 0x61, 0x2e, 0x20, 0x55, 0x74, 0x20, 0x65,
	0x6e, 0x69, 0x6d, 0x20, 0x61, 0x64, 0x20, 0x6d, 0x69, 0x09, 0x00, 0xf2, 0x1a, 0x76, 0x65, 0x6e,
	0x69, 0x61, 0x6d, 0x2c, 0x20, 0x71, 0x75, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x73, 0x74, 0x72, 0x75,
	0x64, 0x20, 0x65, 0x78, 0x65, 0x72, 0x63, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x75,
	0x6c, 0x6c, 0x61, 0x6d, 0x63, 0x6f, 0x5a, 0x00, 0x00, 0x25, 0x00, 0x62, 0x69, 0x73, 0x69, 0x20,
	0x75, 0x74, 0x53, 0x00, 0xf2, 0x01, 0x69, 0x70, 0x20, 0x65, 0x78, 0x20, 0x65, 0x61, 0x20, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x64, 0x6f, 0xc1, 0x00, 0x70, 0x71, 0x75, 0x61, 0x74, 0x2e, 0x20, 0x44,
	0x53, 0x00, 0xa2, 0x61, 0x75, 0x74, 0x65, 0x20, 0x69, 0x72, 0x75, 0x72, 0x65, 0x91, 0x00, 0xf0,
	0x02, 0x20, 0x69, 0x6e, 0x20, 0x72, 0x65, 0x70, 0x72, 0x65, 0x68, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x69, 0x74, 0x11, 0x00, 0xb0, 0x76, 0x6f, 0x6c, 0x75, 0x70, 0x74, 0x61, 0x74, 0x65, 0x20, 0x76,
	0xea, 0x00, 0xa4, 0x20, 0x65, 0x73, 0x73, 0x65, 0x20, 0x63, 0x69, 0x6c, 0x6c, 0x22, 0x01, 0xd0,
	0x65, 0x20, 0x65, 0x75, 0x20, 0x66, 0x75, 0x67, 0x69, 0x61, 0x74, 0x20, 0x6e, 0x91, 0x00, 0xf0,
	0x04, 0x20, 0x70, 0x61, 0x72, 0x69, 0x61, 0x74, 0x75, 0x72, 0x2e, 0x20, 0x45, 0x78, 0x63, 0x65,
	0x70, 0x74, 0x65, 0x75, 0x47, 0x01, 0xf0, 0x04, 0x6e, 0x74, 0x20, 0x6f, 0x63, 0x63, 0x61, 0x65,
	0x63, 0x61, 0x74, 0x20, 0x63, 0x75, 0x70, 0x69, 0x64, 0x61, 0x74, 0x32, 0x00, 0xa0, 0x6f, 0x6e,
	0x20, 0x70, 0x72, 0x6f, 0x69, 0x64, 0x65, 0x6e, 0x46, 0x01, 0x00, 0x2a, 0x01, 0x80, 0x69, 0x6e,
	0x20, 0x63, 0x75, 0x6c, 0x70, 0x61, 0xf8, 0x00, 0xe0, 0x20, 0x6f, 0x66, 0x66, 0x69, 0x63, 0x69,
	0x61, 0x20, 0x64, 0x65, 0x73, 0x65, 0x72, 0x1e, 0x00, 0x40, 0x6d, 0x6f, 0x6c, 0x6c, 0x93, 0x01,
	0x00, 0x21, 0x01, 0xf0, 0x01, 0x69, 0x64, 0x20, 0x65, 0x73, 0x74, 0x20, 0x6c, 0x61, 0x62, 0x6f,
	0x72, 0x75, 0x6d, 0x2e,
	0x0a,
	0x00, 0x00, 0x00, 0x00, // EndMark
	0x59, 0xca, 0x9c, 0x7e, // ContentChecksum
}
var testLoremTXT = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n"

func TestDecompressBlock(t *testing.T) {
	var block = testLoremLZ4[11:432]
	out := make([]byte, len(testLoremTXT))
	n, err := DecompressBlock(block, out)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(testLoremTXT) {
		t.Errorf("Decompressed len mismatch %d != %d", len(testLoremTXT), n)
	}
	if string(out) != testLoremTXT {
		t.Errorf(
			"Deompressed content not match expectations\nExpected\n'%s'\nHave\n'%s'\n",
			testLoremTXT,
			string(out),
		)
	}
}

func TestDecompress(t *testing.T) {
	r := bytes.NewReader(testLoremLZ4)
	w := new(bytes.Buffer)
	d := NewDecompressor()
	err := d.Decompress(r, w)
	if err != nil {
		t.Fatal(err)
	}
	if w.Len() != len(testLoremTXT) {
		t.Errorf("Decompressed len mismatch %d != %d", len(testLoremTXT), w.Len())
	}
	if w.String() != testLoremTXT {
		t.Errorf(
			"Deompressed content not match expectations\nExpected\n'%s'\nHave\n'%s'\n",
			testLoremTXT,
			w.String(),
		)
	}
}

func TestDecompressMultiblock(t *testing.T) {
	f, err := os.Open("war-and-peace.txt.lz4")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f2, err := os.Create("war-and-peace2.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	d := NewDecompressor()
	err = d.Decompress(f, f2)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkDecompress(b *testing.B) {
	d := NewDecompressor()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(testLoremLZ4)
		err := d.Decompress(r, ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompressBlock(b *testing.B) {
	out := make([]byte, len(testLoremTXT))
	d := NewBlockDecoder(testLoremLZ4[11:432], out)
	for i := 0; i < b.N; i++ {
		_, err := d.DecompressBlock(421)
		if err != nil {
			b.Fatal(err)
		}
		d.Reset()
	}
}
