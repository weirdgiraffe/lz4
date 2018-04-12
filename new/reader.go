package new

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/OneOfOne/xxhash"
)

type Reader struct {
	io  io.Reader
	fd  *frameDescriptor
	buf *bytes.Buffer
	in  []byte
	mx  sync.Mutex
}

func NewReader(r io.Reader) *Reader {
	return &Reader{io: r}
}

func (r *Reader) Read(b []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.fd == nil {
		if err := r.initFrame(); err != nil {
			return 0, err
		}
	}

	mustPreserve := 0
	if !r.fd.BlocksIndependent {
		mustPreserve = 64 >> 10
	}

	if r.buf.Len() > mustPreserve {
		return r.buf.Read(b[:len(b)-mustPreserve])
	}

	err := r.decodeBlock()
	if err != nil {
		return 0, err
	}

	return r.buf.Read(b[:len(b)-mustPreserve])
}

// consider not independent blocks
func (r *Reader) decodeBlock() error {
	n, err := readUint32(r.io)
	if err != nil {
		return err
	}

	compressed := (n & 0x80000000) == 0
	n &= 0x7fffffff
	// fmt.Printf("1) block size=%d compressed=%v \n", n, compressed)
	// end mark
	if n == 0 {
		// read the content checksum
		if r.fd.HasContentChecksum {
			_, err := readUint32(r.io)
			if err != nil {
				return err
			}
		}
		return io.EOF
	}

	err = read(r.io, r.in[:n])
	if err != nil {
		return err
	}

	if r.fd.BlocksChecksum {
		chksum, err := readUint32(r.io)
		if err != nil {
			return err
		}
		if xxhash.Checksum32(r.in[:n]) != chksum {
			return fmt.Errorf("decodeBlock: block checksum mismatch")
		}
	}

	if compressed {
		return uncompressBlock(r.in[:n], r.buf)
	}

	_, err = r.buf.Write(r.in[:n])
	return err
}

func (r *Reader) initFrame() error {
	var b [4]byte
	err := read(r.io, b[:])
	if err != nil {
		return err
	}
	if leUint32(b[:]) != 0x184d2204 {
		return fmt.Errorf("Frame: invalid magic")
	}
	r.fd, err = readFrameDescriptor(r.io)
	if err == nil {
		r.in = make([]byte, r.fd.BlockMaxSize)
		r.buf = new(bytes.Buffer)
	}
	return err
}

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

func readUint32(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	err := read(r, b)
	if err != nil {
		return 0, err
	}
	return leUint32(b), nil
}
