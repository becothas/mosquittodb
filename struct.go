package mosquittodb

import (
	"io"
	"os"
)

var (
	Magic = [15]byte{0x00, 0xB5, 0x00, 'm', 'o', 's', 'q', 'u', 'i', 't', 't', 'o', ' ', 'd', 'b'}
)

type Chunk interface {
	Type() ChunkType
}

type Property interface {
}

type DBID uint64

type DB struct {
	file   *os.File
	reader io.Reader
	Header Header
	Config ConfigChunk
}

type Header struct {
	Magic   [15]byte
	CRC     uint32
	Version uint32
}

type ChunkHeader struct {
	Type   ChunkType
	Length uint32
}

type ClientDataChunk struct {
}
