package mosquittodb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type ConfigChunk struct {
	LastStoreID DBID
	Shutdown    uint8
	StoreIDSize uint8
}

func (c ConfigChunk) String() string {
	return fmt.Sprintf("[Config] LastID: %d; Shutdown: %d; DBIdSize: %d", c.LastStoreID, c.Shutdown, c.StoreIDSize)
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
		var err error
		// apparently, this chunk can be padded in the end with a lot of 0s
		cfg.StoreIDSize, err = findStoreIDSize(data)
		if err != nil {
			return err
		}
		cfg.Shutdown = data[cfg.StoreIDSize]
		cfg.LastStoreID, _ = readDBID(breader, cfg.StoreIDSize)
	default:
		err := binary.Read(breader, binary.BigEndian, &cfg.Shutdown)
		if err != nil {
			return err
		}
		err = binary.Read(breader, binary.BigEndian, &cfg.StoreIDSize)
		if err != nil {
			return err
		}
		if len(data) != ((int(cfg.StoreIDSize) & 0xFF) + 2) {
			return ErrUnexpectedConfigurationChunkSize
		}
		cfg.LastStoreID, err = readDBID(breader, cfg.StoreIDSize)
		if err != nil {
			return err
		}
	}
	d.Config = *cfg
	return nil
}
