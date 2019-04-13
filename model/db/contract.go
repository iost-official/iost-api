package db

import (
	"encoding/hex"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
)

type ContractTx struct {
	BID    bson.ObjectId `bson:"_id,omitempty" json:"id"`
	ID     string        `bson:"id"`
	Time   int64         `bson:"time"`
	TxHash string        `bson:"txHash"`
}

type Contract struct {
	ID           string          `bson:"id"`
	Domain       string          `bson:"domain"`
	CreateTime   int64           `bson:"createTime"`
	Creator      string          `bson:"creator"`
	Balance      float64         `bson:"balance"`
	ContractInfo *rpcpb.Contract `bson:"contractInfo"`
}

func NewContract(id string, time int64, creator string) *Contract {
	return &Contract{
		ID:         id,
		CreateTime: time,
		Creator:    creator,
	}
}

func getContractTxQuery(id string) bson.M {
	query := bson.M{
		"id": id,
	}
	return query
}

func GetContractTxByID(id string, start, limit int, ascending bool) ([]*ContractTx, error) {
	contractTxC := GetCollection(CollectionContractTx)

	query := getContractTxQuery(id)
	var contractTxList []*ContractTx
	var sort = "-time"
	if ascending {
		sort = "time"
	}
	err := contractTxC.Find(query).Sort(sort).Skip(start).Limit(limit).All(&contractTxList)
	if err != nil {
		return nil, err
	}
	return contractTxList, nil
}

func GetContractTxByIDAndPos(id, pos string, limit int, ascending bool) ([]*ContractTx, error) {
	d, err := hex.DecodeString(pos)
	if err != nil {
		return nil, err
	}
	contractTxC := GetCollection(CollectionContractTx)

	query := getContractTxQuery(id)
	query["_id"] = bson.M{"$lt": bson.ObjectId(d)}
	var contractTxList []*ContractTx
	var sort = "-_id"
	if ascending {
		sort = "_id"
		query["_id"] = bson.M{"$gt": bson.ObjectId(d)}
	}
	s := time.Now().UnixNano()
	err = contractTxC.Find(query).Sort(sort).Limit(limit).All(&contractTxList)
	log.Printf("ContractTx query cost %d ns, sql: %+v", time.Now().UnixNano()-s, query)
	if err != nil {
		return nil, err
	}
	return contractTxList, nil
}

func GetContractTxNumber(id string) (int, error) {
	contractTxC := GetCollection(CollectionContractTx)

	query := getContractTxQuery(id)
	return contractTxC.Find(query).Count()
}
