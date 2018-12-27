package cron

import (
	"fmt"
	"github.com/iost-official/iost-api/model/blockchain"
	"github.com/iost-official/iost-api/model/db"
	"log"
	"sync"
	"time"
)

func CheckRestart(ws *sync.WaitGroup) {
	defer ws.Done()

	ticker := time.NewTicker(time.Second * 20)

	for range ticker.C {
		// 如果正在更新tx 直接退出
		//if UpdatingTx {
		//	continue
		//}
		// 判断block 的高度
		topHeightInChain, err := blockchain.GetCurrentBlockHeight()
		if err != nil {
			log.Println("updateBlock get topBlk in chain error:", err)
			continue
		}
		topBlkInMongo, err := db.GetTopBlock()
		if err != nil {
			log.Println("updateBlock get topBlk in mongo error:", err)
			continue
		}

		// drop database
		topHeightInMongo := topBlkInMongo.BlockNumber
		if topHeightInChain < topHeightInMongo {
			fmt.Println("drop database")
			db, err := db.GetDb()
			if err != nil {
				continue
			}
			err = db.DropDatabase()
			if err != nil {
				fmt.Println("Drop database error")
			}
		}

	}
}
