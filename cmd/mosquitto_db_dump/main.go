package main

import (
	"flag"
	"fmt"
	"github.com/iotopen/mosquittodb"
	"log"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalln("Usage: mosquitto_db_dump <mosquitto.db>")
	}
	db, err := mosquittodb.Open(args[0])
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("DBVersion:", db.Version())
	hdr := mosquittodb.ChunkHeader{}
	for err := db.ReadChunkHeader(&hdr); err == nil; err = db.ReadChunkHeader(&hdr) {
		switch hdr.Type {
		case mosquittodb.DBChunkCFG:
			cfg := mosquittodb.ConfigChunk{}
			err := db.ReadConfigChunk(&hdr, &cfg)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(cfg)
		case mosquittodb.DBChunkClient:
			client := mosquittodb.ClientChunk{}
			err := db.ReadClientChunk(&hdr, &client)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(client)
		case mosquittodb.DBChunkSub:
			subscription := mosquittodb.SubscriptionChunk{}
			err := db.ReadSubscriptionChunk(&hdr, &subscription)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(subscription)
		case mosquittodb.DBChunkClientMsg:
			msg := mosquittodb.ClientMsgChunk{}
			err := db.ReadClientMsgChunk(&hdr, &msg)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(msg)
		case mosquittodb.DBChunkMsgStore:
			msg := mosquittodb.MsgStoreChunk{}
			err := db.ReadMsgStoreChunk(&hdr, &msg)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(msg)
		case mosquittodb.DBChunkRetain:
			ret := mosquittodb.RetainChunk{}
			err := db.ReadRetainChunk(&hdr, &ret)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(ret)
		default:
			log.Println(fmt.Sprintf("[UnknownChunk](id:%d)", hdr.Type))
			err := db.Skip(&hdr)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	db.Close()
}
