package new

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFrameDescriptor(t *testing.T) {
	var tt = []struct {
		Name       string
		b          []byte
		Error      bool
		Descriptor *frameDescriptor
	}{
		{
			"64K block",
			[]byte{0x40, 0x40, 0xc0},
			false,
			&frameDescriptor{BlockMaxSize: block64KB},
		},
		{
			"256K block",
			[]byte{0x40, 0x50, 0x77},
			false,
			&frameDescriptor{BlockMaxSize: block256KB},
		},
		{
			"1M block",
			[]byte{0x40, 0x60, 0x96},
			false,
			&frameDescriptor{BlockMaxSize: block1MB},
		},
		{
			"4M block",
			[]byte{0x40, 0x70, 0xdf},
			false,
			&frameDescriptor{BlockMaxSize: block4MB},
		},
		{
			"all bits set",
			[]byte{0x7c, 0x40, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xdf},
			false,
			&frameDescriptor{
				BlocksIndependent:  true,
				BlocksChecksum:     true,
				HasContentSize:     true,
				HasContentChecksum: true,
				BlockMaxSize:       block64KB,
				ContentSize:        10,
			},
		},
		{
			"not enough input start",
			[]byte{},
			true,
			nil,
		},
		{
			"not enought input content size",
			[]byte{0x48, 0x40, 0xfd},
			true,
			nil,
		},
		{
			"invalid version bits",
			[]byte{0x00, 0x00, 0x00},
			true,
			nil,
		},
		{
			"invalid reserved bits FLG",
			[]byte{0x43, 0x8f, 0x00},
			true,
			nil,
		},
		{
			"invalid reserved bits BD",
			[]byte{0x40, 0x30, 0xc0},
			true,
			nil,
		},
		{
			"checksum mismatch",
			[]byte{0x40, 0x40, 0xff},
			true,
			nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := bytes.NewReader(tc.b)
			fd, err := readFrameDescriptor(r)
			if err != nil {
				if !tc.Error {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			assert.Equal(t, tc.Descriptor, fd)
		})
	}
}

func TestReader(t *testing.T) {
	r := bytes.NewReader(testLoremLZ4)
	d := NewReader(r)

	text, err := ioutil.ReadAll(d)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testLoremTXT, string(text))
}

func TestReader2(t *testing.T) {
	compressed, err := ioutil.ReadFile("encoded.lz4")
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(compressed)
	r := NewReader(buf)
	if _, err := ioutil.ReadAll(r); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkDecompress(b *testing.B) {
	compressed, err := ioutil.ReadFile("encoded.lz4")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(compressed)
		r := NewReader(buf)
		if _, err := ioutil.ReadAll(r); err != nil {
			b.Fatal(err)
		}
	}
}
