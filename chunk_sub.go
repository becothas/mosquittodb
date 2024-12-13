package mosquittodb

import (
	"encoding/binary"
	"fmt"
)

type SubscriptionChunk struct {
	Identifier uint32
	QoS        uint8
	Options    uint8

	ClientID string
	Topic    string
}

func (s SubscriptionChunk) String() string {
	return fmt.Sprintf("[Sub](q:%d)-'%s': %q", s.QoS, s.ClientID, s.Topic)
}

func (d *DB) readSubscriptionChunkV5(hdr *ChunkHeader, chunk *SubscriptionChunk) error {
	var err error
	pfSub := struct {
		Identifier uint32
		IDLen      uint16
		TopicLen   uint16
		QoS        uint8
		Options    uint8
		Padding    [2]uint8
	}{}
	if err := binary.Read(d.reader, binary.BigEndian, &pfSub); err != nil {
		return err
	}

	chunk.Identifier = pfSub.Identifier
	chunk.QoS = pfSub.QoS
	chunk.Options = pfSub.Options
	if pfSub.IDLen > 0 {
		chunk.ClientID, err = readStringLen(d.reader, pfSub.IDLen)
		if err != nil {
			return err
		}
	}
	if pfSub.TopicLen > 0 {
		chunk.Topic, err = readStringLen(d.reader, pfSub.TopicLen)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) readSubscriptionChunkV234(hdr *ChunkHeader, chunk *SubscriptionChunk) error {
	var err error
	chunk.ClientID, err = readString(d.reader)
	if err != nil {
		return err
	}
	chunk.Topic, err = readString(d.reader)
	if err != nil {
		return err
	}
	if err = binary.Read(d.reader, binary.BigEndian, &chunk.QoS); err != nil {
		return err
	}
	return nil
}

func (d *DB) ReadSubscriptionChunk(hdr *ChunkHeader, chunk *SubscriptionChunk) error {
	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		return d.readSubscriptionChunkV5(hdr, chunk)
	default:
		return d.readSubscriptionChunkV234(hdr, chunk)
	}
}
