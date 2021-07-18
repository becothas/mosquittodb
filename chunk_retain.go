package mosquittodb

import "fmt"

type RetainChunk struct {
	ID DBID
}

func (r RetainChunk) String() string {
	return fmt.Sprintf("[Retained](id:%d)", r.ID)
}

func (d *DB) ReadRetainChunk(hdr *ChunkHeader, chunk *RetainChunk) error {
	if hdr.Type != DBChunkRetain {
		return ErrBadChunkID
	}
	var err error
	chunk.ID, err = readDBID(d.reader, d.Config.DBIDSize)
	if err != nil {
		return err
	}
	return nil
}
