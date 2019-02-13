package db

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	MongoLink     = "mongodb://127.0.0.1:27017"
	MongoUser     = ""
	MongoPassWord = ""
	Db            string
)

const (
	CollectionBlocks     = "blocks"
	CollectionTxs        = "txs"
	CollectionFlatTx     = "flatxs"
	CollectionAccount    = "accounts"
	CollectionAccountTx  = "accountTx"
	CollectionContract   = "contracts"
	CollectionContractTx = "contractTx"
	CollectionTaskCursor = "taskCursors"
	CollectionBlockPay   = "blockPays"
)

func InitConfig() {
	dbConfig := viper.GetStringMapString("mongodb")
	Db = dbConfig["db"]
	MongoUser = dbConfig["username"]
	MongoPassWord = dbConfig["password"]
	MongoLink = fmt.Sprintf("%s:%s", dbConfig["host"], dbConfig["port"])
	fmt.Println("mongolink", Db, MongoLink)
}
