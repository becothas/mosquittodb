package mosquittodb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ConfigChunk struct {
	LastDBID DBID
	Shutdown uint8
	DBIDSize uint8
}

func (c ConfigChunk) String() string {
	return fmt.Sprintf("[Config] LastID: %d; Shutdown: %d; DBIdSize: %d", c.LastDBID, c.Shutdown, c.DBIDSize)
}

func (d *DB) ReadConfigChunk(hdr *ChunkHeader, cfg *ConfigChunk) error {
	if hdr.Type != DBChunkCFG {
		return ErrBadChunkID
	}
	data := make([]byte, hdr.Length)
	_, err := io.ReadFull(d.reader, data)
	if err != nil {
		return err
	}

	breader := bytes.NewReader(data)
	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		cfg.DBIDSize = data[len(data)-1]
		if len(data) != ((int(cfg.DBIDSize) & 0xFF) + 2) {
			return ErrUnexpectedConfigurationChunkSize
		}
		cfg.Shutdown = data[len(data)-2]
		cfg.LastDBID, _ = readDBID(breader, cfg.DBIDSize)
	default:
		err := binary.Read(breader, binary.BigEndian, &cfg.Shutdown)
		if err != nil {
			return err
		}
		err = binary.Read(breader, binary.BigEndian, &cfg.DBIDSize)
		if err != nil {
			return err
		}
		if len(data) != ((int(cfg.DBIDSize) & 0xFF) + 2) {
			return ErrUnexpectedConfigurationChunkSize
		}
		cfg.LastDBID, err = readDBID(breader, cfg.DBIDSize)
		if err != nil {
			return err
		}
	}
	d.Config = *cfg
	return nil
}
