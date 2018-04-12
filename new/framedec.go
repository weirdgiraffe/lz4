package new

import (
	"fmt"
	"io"

	"github.com/OneOfOne/xxhash"
)

const (
	block64KB  = 65536
	block256KB = 262144
	block1MB   = 1048576
	block4MB   = 4194304
)

type frameDescriptor struct {
	BlocksIndependent  bool
	BlocksChecksum     bool
	HasContentSize     bool
	HasContentChecksum bool
	BlockMaxSize       int
	ContentSize        uint64
}

func readFrameDescriptor(r io.Reader) (*frameDescriptor, error) {
	var b [11]byte
	// read either FLG,BD and HC bytes or FLG,BD and first byte of content size
	err := read(r, b[:3])
	if err != nil {
		return nil, err
	}
	if b[0]&0xc3 != 0x40 {
		return nil, fmt.Errorf("invalid version or reseved bits in FLG byte")
	}
	if b[1]&0x8f != 0x00 {
		return nil, fmt.Errorf("invalid reseved bits in BD byte")
	}
	if b[1]&0xcf != 0x40 {
		return nil, fmt.Errorf("invalid max block size in BD byte")
	}
	fd := &frameDescriptor{
		BlocksIndependent:  b[0]&0x20 != 0,
		BlocksChecksum:     b[0]&0x10 != 0,
		HasContentSize:     b[0]&0x08 != 0,
		HasContentChecksum: b[0]&0x04 != 0,
		// math here is the following:
		// 0XXX0000 is the bitmask for a max block size
		// 0x40 -> 64K  = 2^16  | 2^(4 + 2 + 10) = 2^(2*4 + 8)
		// 0x50 -> 256K = 2^18  | 2^(5 + 3 + 10) = 2^(2*5 + 8)
		// 0x60 -> 1M   = 2^20  | 2^(6 + 4 + 10) = 2^(2*6 + 8)
		// 0x70 -> 4M   = 2^22  | 2^(7 + 5 + 10) = 2^(2*7 + 8)
		BlockMaxSize: 1 << (2*((b[1]&0x70)>>4) + 8),
	}
	hc := 2
	// read rest bytes of frame descriptor
	if fd.HasContentSize {
		err = read(r, b[3:11])
		if err != nil {
			return nil, err
		}
		fd.ContentSize = leUint64(b[2:10])
		hc = 10
	}
	chksum := byte((xxhash.Checksum32(b[:hc]) >> 8) & 0xff)
	if chksum != b[hc] {
		err = fmt.Errorf(
			"FrameDesc: checksum mismatch %02x != %02x",
			chksum, b[hc],
		)
		return nil, err
	}
	return fd, nil
}

func read(r io.Reader, b []byte) error {
	l := 0
	for l < len(b) {
		n, err := r.Read(b[l:])
		if err != nil {
			return err
		}
		l += n
	}
	return nil
}
