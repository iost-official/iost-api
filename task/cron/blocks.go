package cron

import (
	"log"
	"sync"
	"time"

	"github.com/iost-official/iost-api/model/blockchain"
	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
	"github.com/iost-official/iost-api/model/db"
)

func UpdateBlocks(ws *sync.WaitGroup) {
	defer ws.Done()

	blockChannel := make(chan *rpcpb.Block, 10)
	go insertBlock(blockChannel)

	ticker := time.NewTicker(time.Second)

	var topHeightInMongo int64
	for range ticker.C {
		topBlkInMongo, err := db.GetTopBlock()
		if err != nil {
			log.Println("updateBlock get topBlk in mongo error:", err)
			if err.Error() != "not found" {
				continue
			} else {
				topHeightInMongo = 0
				break
			}
		}
		topHeightInMongo = topBlkInMongo.Number + 1
		log.Println("Got Top Block From Mongo With Number: ", topHeightInMongo)
		break
	}

	for {
		blockRspn, err := blockchain.GetBlockByNum(topHeightInMongo, true)
		if err != nil {
			log.Println("Download block", topHeightInMongo, "error:", err)
			time.Sleep(time.Second)
			continue
		}
		if blockRspn.Status == rpcpb.BlockResponse_PENDING {
			log.Println("Download block", topHeightInMongo, "Pending")
			time.Sleep(time.Second)
			continue
		}
		blockChannel <- blockRspn.Block
		log.Println("Download block", topHeightInMongo, " Succ!")
		topHeightInMongo++
	}
}

func insertBlock(blockChannel chan *rpcpb.Block) {
	collection, err := db.GetCollection(db.CollectionBlocks)
	if err != nil {
		log.Println("can not get blocks collection when update", err)
		return
	}

	for {
		select {
		case b := <-blockChannel:
			txs := b.Transactions

			db.ProcessTxs(txs)

			b.Transactions = make([]*rpcpb.Transaction, 0)
			err = collection.Insert(*b)

			if err != nil {
				log.Println("updateBlock insert mongo error:", err)
			}
		default:

		}
	}
}
