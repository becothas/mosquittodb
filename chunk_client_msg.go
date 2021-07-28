package mosquittodb

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type ClientMsgChunk struct {
	StoreID        DBID
	MID            uint16
	QoS            uint8
	State          uint8
	Retain         bool
	Dup            bool
	Direction      uint8
	ClientID       string
	SubscriptionID uint32 // from MQTT_PROP_SUBSCRIPTION_IDENTIFIER
	Properties     []Property
}

func (c ClientMsgChunk) String() string {
	return fmt.Sprintf("[ClientMsg](q:%d, ret:%v, dup:%v, dir:%d)-'%s' MID:%d State:%d",
		c.QoS, c.Retain, c.Dup, c.Direction, c.ClientID, c.MID, c.State)
}

func (d *DB) readClientMsgChunkV56(hdr *ChunkHeader, chunk *ClientMsgChunk) error {
	length := hdr.Length
	var err error
	chunk.StoreID, err = readDBID(d.reader, d.Config.StoreIDSize)
	if err != nil {
		return err
	}

	pfClientMsg := struct {
		MID       uint16
		IDLen     uint16
		QoS       uint8
		State     uint8
		RetainDup uint8
		Direction uint8
	}{}

	if err := binary.Read(d.reader, binary.BigEndian, &pfClientMsg); err != nil {
		return err
	}

	chunk.MID = pfClientMsg.MID
	chunk.QoS = pfClientMsg.QoS
	chunk.State = pfClientMsg.State
	chunk.Retain = (pfClientMsg.RetainDup & 0xF0) > 0
	chunk.Dup = (pfClientMsg.RetainDup & 0x0F) > 0
	chunk.Direction = pfClientMsg.Direction

	length -= (uint32(d.Config.StoreIDSize) & 0xFF) +
		uint32(unsafe.Sizeof(pfClientMsg)) +
		(uint32(pfClientMsg.IDLen) & 0xFFFF)

	chunk.ClientID, err = readStringLen(d.reader, pfClientMsg.IDLen)
	if err != nil {
		return err
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

func (d *DB) readClientMsgChunkV234(hdr *ChunkHeader, chunk *ClientMsgChunk) error {
	var err error
	retain, dup := uint8(0), uint8(0)

	chunk.ClientID, err = readString(d.reader)
	if err != nil {
		return err
	}

	chunk.StoreID, err = readDBID(d.reader, d.Config.StoreIDSize)
	if err != nil {
		return err
	}

	if err := binary.Read(d.reader, binary.BigEndian, &chunk.MID); err != nil {
		return err
	}
	if err := binary.Read(d.reader, binary.BigEndian, &chunk.QoS); err != nil {
		return err
	}

	if err := binary.Read(d.reader, binary.BigEndian, &retain); err != nil {
		return err
	}
	if retain > 0 {
		chunk.Retain = true
	}

	if err := binary.Read(d.reader, binary.BigEndian, &chunk.Direction); err != nil {
		return err
	}
	if err := binary.Read(d.reader, binary.BigEndian, &chunk.State); err != nil {
		return err
	}
	if err := binary.Read(d.reader, binary.BigEndian, &dup); err != nil {
		return err
	}
	if dup > 0 {
		chunk.Dup = true
	}
	return nil
}

func (d *DB) ReadClientMsgChunk(hdr *ChunkHeader, chunk *ClientMsgChunk) error {
	if hdr.Type != DBChunkClientMsg {
		return ErrBadChunkID
	}
	switch d.Version() {
	case MosqDbVersion5, MosqDbVersion6:
		return d.readClientMsgChunkV56(hdr, chunk)
	default:
		return d.readClientMsgChunkV234(hdr, chunk)
	}
}
