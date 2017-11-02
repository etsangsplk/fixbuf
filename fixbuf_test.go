package fixbuf

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

// typeA is automatically encoded
type typeA [3]byte

// typeB is encoded via a custom marshaller
type typeB [2]int16

func (b typeB) MarshalTo(w io.Writer) (int, error) {
	// Just to test special marshalling, make first element little endian
	// and the second element big.
	binary.Write(w, binary.LittleEndian, &b[0])
	binary.Write(w, binary.BigEndian, &b[1])
	return 4, nil
}

func (b typeB) UnmarshalFrom(r io.Reader) (int, error) {
	// Just to test special marshalling, make first element little endian
	// and the second element big.
	binary.Read(r, binary.LittleEndian, &b[0])
	binary.Read(r, binary.BigEndian, &b[1])
	return 4, nil
}

var _ Marshalling = typeB{}

type testStruct struct {
	A typeA
	B typeB
}

func TestRoundtrip(t *testing.T) {
	ones := typeA{}
	for i := range ones {
		ones[i] = 1
	}
	twos := typeB{}
	for i := range twos {
		twos[i] = 2
	}
	x := testStruct{A: ones, B: twos}
	enc := NewBinaryEncoding(nil)

	buf := &bytes.Buffer{}
	enc.Write(buf, x, x)
	exp := []byte{0x1, 0x1, 0x1, 0x2, 0x0, 0x0, 0x2, 0x1, 0x1, 0x1, 0x2, 0x0, 0x0, 0x2}
	if !bytes.Equal(exp, buf.Bytes()) {
		t.Fatalf("expected %v, got %v", exp, buf.Bytes())
	}

	rbuf := bytes.NewBuffer(buf.Bytes())
	var y testStruct
	enc.Read(rbuf, &y)

	buf2 := &bytes.Buffer{}
	enc.Write(buf2, y, y)

	// This is failing because of issue #2 right now.
	if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
		t.Fatalf("expected %#v, got %#v", buf.Bytes(), buf2.Bytes())
	}
}
