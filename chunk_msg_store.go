package mosquittodb

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
	"unsafe"
)

type MsgStoreChunk struct {
	ID         DBID
	ExpiryTime time.Time
	SourceMID  uint16
	SourcePort uint16
	QoS        uint8
	Retain     bool

	Payload        []byte
	SourceID       string
	SourceUsername string
	Topic          string
	Properties     []Property
}

func (m MsgStoreChunk) String() string {
	return fmt.Sprintf("[Store](q:%d, ret:%v)-'%s': %q", m.QoS, m.Retain, m.SourceID, m.Topic)
}

func (d *DB) readMsgStoreChunkV56(hdr *ChunkHeader, chunk *MsgStoreChunk) error {
	var err error
	length := hdr.Length
	msgStoreLengths := struct {
		ExpiryTime        int64
		PayloadLen        uint32
		SourceMID         uint16
		SourceIDLen       uint16
		SourceUsernameLen uint16
		TopicLen          uint16
		SourcePort        uint16
		Qos               uint8
		Retain            uint8
	}{}
	if err := binary.Read(d.reader, binary.BigEndian, &msgStoreLengths); err != nil {
		return err
	}
	chunk.ExpiryTime = time.Unix(msgStoreLengths.ExpiryTime, 0)
	chunk.SourceMID = msgStoreLengths.SourceMID
	chunk.SourcePort = msgStoreLengths.SourcePort
	chunk.QoS = msgStoreLengths.Qos
	if msgStoreLengths.Retain > 0 {
		chunk.Retain = true
	}

	length -= (uint32(d.Config.DBIDSize) & 0xFF) +
		uint32(unsafe.Sizeof(msgStoreLengths)) +
		msgStoreLengths.PayloadLen +
		(uint32(msgStoreLengths.SourceIDLen) & 0XFFFF) +
		(uint32(msgStoreLengths.SourceUsernameLen) & 0xFFFF) +
		(uint32(msgStoreLengths.TopicLen) & 0xFFFF)

	if msgStoreLengths.SourceIDLen > 0 {
		chunk.SourceID, err = readStringLen(d.reader, msgStoreLengths.SourceIDLen)
		if err != nil {
			return err
		}
	}
	if msgStoreLengths.SourceUsernameLen > 0 {
		chunk.SourceUsername, err = readStringLen(d.reader, msgStoreLengths.SourceUsernameLen)
		if err != nil {
			return err
		}
	}
	if msgStoreLengths.TopicLen > 0 {
		chunk.Topic, err = readStringLen(d.reader, msgStoreLengths.TopicLen)
		if err != nil {
			return err
		}
	}
	chunk.Payload = make([]byte, msgStoreLengths.PayloadLen)
	if msgStoreLengths.PayloadLen > 0 {
		_, err = io.ReadFull(d.reader, chunk.Payload)
		if err != nil {
			return err
		}
	}
	if length > 0 {
		// Todo: Read all properties
		data := make([]byte, length)
		_, err = io.ReadFull(d.reader, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) readMsgStoreChunkV234(hdr *ChunkHeader, chunk *MsgStoreChunk) error {
	var err error
	chunk.SourceID, err = readString(d.reader)
	if err != nil {
		return err
	}
	if d.Version() == MosqDbVersion4 {
		chunk.SourceUsername, err = readString(d.reader)
		if err != nil {
			return err
		}
		if err := binary.Read(d.reader, binary.BigEndian, &chunk.SourcePort); err != nil {
			return err
		}
	}
	if err := binary.Read(d.reader, binary.BigEndian, &chunk.SourceMID); err != nil {
		return err
	}

	// This is the mid - don't need it
	dontNeedMid := uint16(0)
	if err := binary.Read(d.reader, binary.BigEndian, &dontNeedMid); err != nil {
		return err
	}

	chunk.Topic, err = readString(d.reader)
	if err != nil {
		return err
	}
	retain := uint8(0)
	if err := binary.Read(d.reader, binary.BigEndian, &chunk.QoS); err != nil {
		return err
	}
	if err := binary.Read(d.reader, binary.BigEndian, &retain); err != nil {
		return err
	}
	if retain > 0 {
		chunk.Retain = true
	}

	payloadLen := uint32(0)
	if err := binary.Read(d.reader, binary.BigEndian, &payloadLen); err != nil {
		return err
	}
	chunk.Payload = make([]byte, payloadLen)
	if payloadLen > 0 {
		_, err = io.ReadFull(d.reader, chunk.Payload)
	}
	return nil
}

func (d *DB) ReadMsgStoreChunk(hdr *ChunkHeader, chunk *MsgStoreChunk) error {
	if hdr.Type != DBChunkMsgStore {
		return ErrBadChunkID
	}
	var err error
	chunk.ID, err = readDBID(d.reader, d.Config.DBIDSize)
	if err != nil {
		return err
	}
	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		return d.readMsgStoreChunkV56(hdr, chunk)
	default:
		return d.readMsgStoreChunkV234(hdr, chunk)
	}
}
