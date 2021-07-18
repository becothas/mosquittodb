package mosquittodb

const (
	MosqDbVersion  = 6
	MosqDbVersion2 = 2
	MosqDbVersion3 = 3
	MosqDbVersion4 = 4
	MosqDbVersion5 = 5
	MosqDbVersion6 = 6
)

type ChunkType uint32

const (
	DBChunkCFG = ChunkType(iota + 1)
	DBChunkMsgStore
	DBChunkClientMsg
	DBChunkRetain
	DBChunkSub
	DBChunkClient
)

const (
	MQTTMaxPayloadLen = 268435455
)
