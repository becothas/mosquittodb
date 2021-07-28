package mosquittodb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

func Open(path string) (*DB, error) {
	var err error
	res := &DB{}
	res.file, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	res.reader = bufio.NewReader(res.file)

	err = readDBHeader(res.reader, &res.Header)
	if err != nil {
		_ = res.file.Close()
		return nil, err
	}

	if !bytecmp(Magic[:], res.Header.Magic[:]) {
		_ = res.file.Close()
		return nil, ErrBadMagic
	}

	if res.Header.Version != MosqDbVersion {
		switch res.Header.Version {
		case MosqDbVersion2:
			// Addition of disconnect_t to client chunk in v3.
		case MosqDbVersion3:
			// Addition of source_username and source_port to msg_store chunk in v4, v1.5.6
		case MosqDbVersion4:
		case MosqDbVersion5:
			// Addition of username and listener_port to client chunk in v6
		default:
			_ = res.file.Close()
			return nil, errors.New(fmt.Sprintf("unsupported database version (%d)", res.Header.Version))
		}
	}

	return res, nil
}

func (d *DB) Close() {
	_ = d.file.Close()
}

func (d *DB) Version() uint32 {
	return d.Header.Version
}

func (d *DB) ReadChunkHeader(hdr *ChunkHeader) error {
	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		err := binary.Read(d.reader, binary.BigEndian, hdr)
		if err != nil {
			return err
		}
	default:
		var i16tmp uint16
		err := binary.Read(d.reader, binary.BigEndian, &i16tmp)
		if err != nil {
			return err
		}
		err = binary.Read(d.reader, binary.BigEndian, &hdr.Length)
		if err != nil {
			return err
		}
		hdr.Type = ChunkType(i16tmp)
	}
	return nil
}

func (d *DB) Skip(hdr *ChunkHeader) error {
	null := make([]byte, hdr.Length)
	_, err := io.ReadFull(d.reader, null)
	return err
}
