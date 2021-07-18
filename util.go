package mosquittodb

import (
	"encoding/binary"
	"io"
	"math/big"
)

func bytecmp(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i, x := range b1 {
		if x != b2[i] {
			return false
		}
	}
	return true
}

func readDBHeader(r io.Reader, header *Header) error {
	err := binary.Read(r, binary.BigEndian, header)
	return err
}

func readDBID(r io.Reader, size uint8) (DBID, error) {
	value := big.NewInt(0)
	if size == 0 {
		size = 8
	}
	dbidData := make([]byte, size)
	_, err := io.ReadFull(r, dbidData)
	if err != nil {
		return DBID(value.Uint64()), err
	}
	value.SetBytes(dbidData)
	return DBID(value.Uint64()), nil
}

func readStringLen(r io.Reader, size uint16) (string, error) {
	data := make([]byte, size)
	_, err := io.ReadFull(r, data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readString(r io.Reader) (string, error) {
	size := uint16(0)
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return "", err
	}
	return readStringLen(r, size)
}
