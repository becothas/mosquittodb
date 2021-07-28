package mosquittodb

import "fmt"

type RetainChunk struct {
	StoreID DBID
}

func (r RetainChunk) String() string {
	return fmt.Sprintf("[Retained](id:%d)", r.StoreID)
}

func (d *DB) ReadRetainChunk(hdr *ChunkHeader, chunk *RetainChunk) error {
	if hdr.Type != DBChunkRetain {
		return ErrBadChunkID
	}
	var err error
	chunk.StoreID, err = readDBID(d.reader, d.Config.StoreIDSize)
	if err != nil {
		return err
	}
	return nil
}
