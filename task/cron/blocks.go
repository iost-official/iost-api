package cron

import (
	"github.com/iost-official/go-iost/core/block"
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
	b := <-blockChannel
	ticker := time.NewTicker(time.Second)

	var topHeightInMongo int64
	for range ticker.C {
	topBlkInMongo, err := db.GetTopBlock()
	if err != nil {
		log.Println("updateBlock get topBlk in mongo error:", err)
		if err.Error() != "not found" {
			continue
		}
	} else {
		topHeightInMongo = topBlkInMongo.BlockNumber + 1
		break
	}
	}

	for {
		blockRspn, err := blockchain.GetBlockByNum(topHeightInMongo)
		if nil != err {
			log.Println("Download block", topHeightInMongo, " error", err)
			time.Sleep(time.Second)
			continue
		}
		if blockRspn.BlockResponse_Status == rpcpb.BlockResponse_PENDING {
			log.Println("Download block", topHeightInMongo, " Pending")
			time.Sleep(time.Second())
			continue
		}
		blockChannel <- blockRspn.Block
		log.Println("Download block", topHeightInMongo, " Succ!")
	}
}

func InsertBlock(blockChannel chan *rpcpb.Block) error{
	collection, err := db.GetCollection("block")
	if nil != err {
		log.Println("can not get blocks collection when update", err)
		return
	}

	select {
	case b := <- blockChannel:
		txs := b.Transactions
		db.ProcessTxs(txs)
		
		b.Transactions = []*rpcpb.Transaction
		err = collection.Insert(&block)

		if err != nil {
			log.Println("updateBlock insert mongo error:", err)

			err := recordFailedUpdateBlock(topHeightInMongo, fBlockCollection)
			if nil != err {
				log.Println("UpdateBlock record sync failed block error", err)
		}

	}
}

func ProcessFailedSyncBlocks(ws *sync.WaitGroup) {
	defer ws.Done()

	collection, err := db.GetCollection(db.CollectionBlocks)
	if nil != err {
		log.Println("can not get blocks collection when update", err)
		return
	}

	fBlockCollection, err := db.GetCollection(db.CollectionFBlocks)
	if nil != err {
		log.Println("Process failed sync blocks get f blocks collection error", err)
		return
	}

	tmpTxCollection, err := db.GetCollection(db.CollectionTmpTxs)

	if nil != err {
		log.Println("Process failed sync blocks get txs collection error", err)
		return
	}

	query := bson.M{
		"processed": false,
		"retryTimes": bson.M{
			"$lte": 5,
		},
	}

	ticker := time.NewTicker(time.Second * 2)

	for range ticker.C {
		var fBlockList = make([]*db.FailBlock, 0)
		fBlockCollection.Find(query).Sort("blockNumber").All(&fBlockList)
		for _, fBlock := range fBlockList {
			block, txHashes, err := db.GetBlockInfoByNum(fBlock.BlockNumber)

			if err != nil {
				log.Println("Process failed blocks rpc call error:", err)
				fBlock.RetryTimes++
				fBlockCollection.Update(bson.M{"blockNumber": fBlock.BlockNumber}, bson.M{"$set": bson.M{
					"retryTimes": fBlock.RetryTimes,
				}})
				continue
			}

			count, err := collection.Find(bson.M{"blockNumber": fBlock.BlockNumber}).Count()

			if nil != err {
				log.Println("Process failed blocks count error", err)
				continue
			}

			if count == 0 {
				err = collection.Insert(block)

				if nil != err {
					log.Println("Process failed blocks insert block error", err)
					fBlock.RetryTimes++
					fBlockCollection.Update(bson.M{"blockNumber": fBlock.BlockNumber}, bson.M{"$set": bson.M{
						"retryTimes": fBlock.RetryTimes,
					}})
					continue
				}
			}

			if nil != txHashes {
				txs := make([]interface{}, len(*txHashes))
				for index, txHash := range *txHashes {
					txs[index] = db.TmpTx{
						Hash:        txHash,
						BlockNumber: fBlock.BlockNumber,
						Mark:        fBlock.BlockNumber % 2,
					}
				}
				err := tmpTxCollection.Insert(txs...)
				if nil != err {
					fBlock.RetryTimes++
					fBlockCollection.Update(bson.M{"blockNumber": fBlock.BlockNumber}, bson.M{"$set": bson.M{
						"retryTimes": fBlock.RetryTimes,
					}})
					log.Println("UpdateBlock2 insert txs error", err)
					continue
				}
			}
			fBlockCollection.Update(bson.M{"blockNumber": fBlock.BlockNumber}, bson.M{"$set": bson.M{
				"processed": true,
			}})
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
