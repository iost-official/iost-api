package cron

import (
	"log"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
			continue
		}

		topHeightInMongo = topBlkInMongo.BlockNumber + 1
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
	}
}

func insertBlock(blockChannel chan *rpcpb.Block) {
	collection, err := db.GetCollection("block")
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
			err = collection.Insert(&b)

			if err != nil {
				log.Println("updateBlock insert mongo error:", err)
			}
		default:

		}
	}
}

func UpdateBlockPay(wg *sync.WaitGroup) {
	defer wg.Done()

	txnC, err := db.GetCollection(db.CollectionTxs)
	if err != nil {
		log.Println("UpdateBlockCost get collection error:", err)
		return
	}

	blkPC, err := db.GetCollection(db.CollectionBlockPay)
	if err != nil {
		log.Println("UpdateBlockCost get collection error:", err)
		return
	}

	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		var topHeightInPay int64
		topPay, err := db.GetTopBlockPay()
		if err != nil {
			if err.Error() != "not found" {
				continue
			}
		} else {
			topHeightInPay = topPay.Height + 1
		}

		queryPip := []bson.M{
			{
				"$match": bson.M{
					"blockNumber": bson.M{
						"$gte": topHeightInPay,
					},
					"time": bson.M{
						"$ne": 0,
					},
				},
			},
			{
				"$group": bson.M{
					"_id": "$blockNumber",
					"avggasprice": bson.M{
						"$avg": "$gasPrice",
					},
					"totalgaslimit": bson.M{
						"$sum": "$gasLimit",
					},
				},
			},
		}

		var payList []*db.BlockPay
		err = txnC.Pipe(queryPip).All(&payList)
		if err != nil {
			log.Println("UpdateBlockPay pipeline error:", err)
			continue
		}

		for _, pay := range payList {
			selector := bson.M{
				"_id": pay.Height,
			}
			blkPC.Upsert(selector, pay)
			log.Println("UpdateBlockPay block:", pay.Height, "inserted")
		}
	}
}

func recordFailedUpdateBlock(blockNumber int64, fBlockCollection *mgo.Collection) error {
	fBlock := db.FailBlock{
		BlockNumber: blockNumber,
		RetryTimes:  0,
		Processed:   false,
	}

	err := fBlockCollection.Insert(fBlock)
	return err
}
