package mosquittodb

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type ClientChunk struct {
	SessionExpiryTime     time.Time
	SessionExpiryInterval uint32
	LastMID               uint16
	ListenerPort          uint16

	ClientID string
	Username string
}

func (c ClientChunk) String() string {
	return fmt.Sprintf("[Client](usr:%q)-'%s' @%d", c.Username, c.ClientID, c.ListenerPort)
}

func (d *DB) readClientChunkV56(hdr *ChunkHeader, chunk *ClientChunk) error {
	var err error
	switch d.Version() {
	case MosqDbVersion6:

		pfClient := struct {
			SessionExpiryTime     int64
			SessionExpiryInterval uint32
			LastMID               uint16
			IDLen                 uint16
			ListenerPort          uint16
			UsernameLen           uint16
			Padding               uint32
		}{}

		if err := binary.Read(d.reader, binary.BigEndian, &pfClient); err != nil {
			return err
		}
		chunk.SessionExpiryTime = time.Unix(pfClient.SessionExpiryTime, 0)
		chunk.SessionExpiryInterval = pfClient.SessionExpiryInterval
		chunk.LastMID = pfClient.LastMID
		chunk.ListenerPort = pfClient.ListenerPort

		chunk.ClientID, err = readStringLen(d.reader, pfClient.IDLen)
		if err != nil {
			return err
		}
		if pfClient.UsernameLen > 0 {
			chunk.Username, err = readStringLen(d.reader, pfClient.UsernameLen)
			if err != nil {
				return err
			}
		}

	case MosqDbVersion5:
		pfClient := struct {
			SessionExpiryTime     int64
			SessionExpiryInterval uint32
			LastMID               uint16
			IDLen                 uint16
		}{}
		if err := binary.Read(d.reader, binary.BigEndian, &pfClient); err != nil {
			return err
		}
		chunk.SessionExpiryTime = time.Unix(pfClient.SessionExpiryTime, 0)
		chunk.SessionExpiryInterval = pfClient.SessionExpiryInterval
		chunk.LastMID = pfClient.LastMID

		chunk.ClientID, err = readStringLen(d.reader, pfClient.IDLen)
		if err != nil {
			return err
		}
	default:
		panic("Unsupported database version")
	}
	return nil
}

func (d *DB) readClientChunkV234(hdr *ChunkHeader, chunk *ClientChunk) error {
	var err error
	chunk.ClientID, err = readString(d.reader)
	if err != nil {
		return err
	}

	if err := binary.Read(d.reader, binary.BigEndian, &chunk.LastMID); err != nil {
		return err
	}

	nRead := 2 + len(chunk.ClientID) + 2

	if d.Version() != MosqDbVersion2 {
		// It looks like there is a time saved in the database for version 3 and 4, but it's never used.
		// We still need to read it.
		diff := hdr.Length - uint32(nRead)
		tmp := make([]byte, diff)
		_, err = io.ReadFull(d.reader, tmp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) ReadClientChunk(hdr *ChunkHeader, chunk *ClientChunk) error {
	if hdr.Type != DBChunkClient {
		return ErrBadChunkID
	}

	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		return d.readClientChunkV56(hdr, chunk)
	default:
		return d.readClientChunkV234(hdr, chunk)
	}
}
