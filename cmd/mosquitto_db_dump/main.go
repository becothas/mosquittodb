package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/iotopen/mosquittodb"
	"os"
)

func assert(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	var err error
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		assert(errors.New("usage: mosquitto_db_dump <mosquitto.db>"))
	}
	db, err := mosquittodb.Open(args[0])
	assert(err)
	defer db.Close()
	fmt.Println("DBVersion:", db.Version())
	hdr := mosquittodb.ChunkHeader{}
	for err = db.ReadChunkHeader(&hdr); err == nil; err = db.ReadChunkHeader(&hdr) {
		switch hdr.Type {
		case mosquittodb.DBChunkCFG:
			cfg := mosquittodb.ConfigChunk{}
			err := db.ReadConfigChunk(&hdr, &cfg)
			assert(err)
			fmt.Println(cfg)
		case mosquittodb.DBChunkClient:
			client := mosquittodb.ClientChunk{}
			err := db.ReadClientChunk(&hdr, &client)
			assert(err)
			fmt.Println(client)
		case mosquittodb.DBChunkSub:
			subscription := mosquittodb.SubscriptionChunk{}
			err := db.ReadSubscriptionChunk(&hdr, &subscription)
			assert(err)
			fmt.Println(subscription)
		case mosquittodb.DBChunkClientMsg:
			msg := mosquittodb.ClientMsgChunk{}
			err := db.ReadClientMsgChunk(&hdr, &msg)
			assert(err)
			fmt.Println(msg)
		case mosquittodb.DBChunkMsgStore:
			msg := mosquittodb.MsgStoreChunk{}
			err := db.ReadMsgStoreChunk(&hdr, &msg)
			assert(err)
			fmt.Println(msg)
		case mosquittodb.DBChunkRetain:
			ret := mosquittodb.RetainChunk{}
			err := db.ReadRetainChunk(&hdr, &ret)
			assert(err)
			fmt.Println(ret)
		default:
			fmt.Println(fmt.Sprintf("[UnknownChunk](id:%d)", hdr.Type))
			err := db.Skip(&hdr)
			assert(err)
		}
	}
}
