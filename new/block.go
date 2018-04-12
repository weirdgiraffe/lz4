package new

import (
	"bytes"
	"fmt"
)

func uncompressBlock(in []byte, out *bytes.Buffer) (err error) {
	var i, ll, ml, mo int
	// fmt.Printf("- 0) copied %s ...\n", hex.EncodeToString(in[:32]))
	// fmt.Printf("- 1) uncompress block of size %d with %d bytes in out buffer\n", len(in), out.Len())
	for {
		// read the token and set initial length for literals and match
		ll, ml = int(in[i]&0xf0>>4), int(in[i]&0x0f)+4
		// fmt.Printf("- - 1) ll=%d ml=%d\n", ll, ml)
		i++
		// read rest of the literals length
		if ll == 15 {
			for ; i < len(in) && in[i] == 255; i++ {
				ll += 255
			}
			ll += int(in[i])
			i++
		}
		// fmt.Printf("- - 2) ll=%d\n", ll)
		// copy literals
		if ll != 0 {
			_, err = out.Write(in[i : i+ll])
			if err != nil {
				return err
			}
			i += ll
		}
		// fmt.Printf("- - 3) copied %d, now we have %d in out\n", ll, out.Len())
		// check if we have processed the whole block
		if i == len(in) {
			return nil
		}
		// read match offset
		mo = int(in[i]) | int(in[i+1])<<8
		// fmt.Printf("- - 3) %02x%02x\n", in[i], in[i+1])
		i += 2
		// fmt.Printf("- - 4) match offset %d out of %d\n", mo, out.Len())
		if mo == 0 || mo > out.Len() {
			return fmt.Errorf("invalid match offset")
		}
		mo = out.Len() - mo
		// fmt.Printf("- - 5) offset adjusted to %d\n", mo)
		// read rest of the match length
		if ml == 19 {
			for ; i < len(in) && in[i] == 255; i++ {
				ml += 255
			}
			ml += int(in[i])
			i++
		}
		// fmt.Printf("- - 6) ml=%d\n", ml)
		// copy match
		//fmt.Printf("len=%d mo=%d ml=%d mo+ml=%d\n", out.Len(), mo, ml, mo+ml)
		_, err = out.Write(out.Bytes()[mo : mo+ml])
		if err != nil {
			return err
		}
	}
	return nil
}
