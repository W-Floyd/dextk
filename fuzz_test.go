package dextk

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// testdata/probe.dex is testdata/Probe.java built by
//
//	javac --release 11 -d classes Probe.java
//	d8 --min-api 26 --output . classes/com/example/*.class
func readAll(t testing.TB, b []byte) {
	r, err := Read(bytes.NewReader(b))
	if err != nil {
		return
	}
	for it := r.StringIter(); it.HasNext(); {
		if _, err := it.Next(); err != nil {
			break
		}
	}
	for it := r.ClassIter(); it.HasNext(); {
		c, err := it.Next()
		if err != nil {
			break
		}
		for _, ms := range [][]MethodNode{c.DirectMethods, c.VirtualMethods} {
			for _, m := range ms {
				if m.CodeOff != 0 {
					r.ReadCodeAndParse(m.CodeOff)
				}
			}
		}
	}
}

// A count read from the file must not size an allocation past the file.
func TestStringSizePastFile(t *testing.T) {
	b, err := os.ReadFile("testdata/probe.dex")
	if err != nil {
		t.Fatal(err)
	}
	sid := binary.LittleEndian.Uint32(b[0x3c:])
	off := binary.LittleEndian.Uint32(b[sid:])
	copy(b[off:], []byte{0xff, 0xff, 0xff, 0xff, 0x0f}) // ULEB128 0xffffffff
	r, err := Read(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ReadString(0); err != ErrOutOfBounds {
		t.Errorf("got %v, want ErrOutOfBounds", err)
	}
}

func FuzzRead(f *testing.F) {
	b, err := os.ReadFile("testdata/probe.dex")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		defer func() { recover() }() // panics are a separate matter; this hunts hangs and huge allocations
		readAll(t, b)
	})
}
