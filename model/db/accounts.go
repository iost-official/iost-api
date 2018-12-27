package db

import (
	"log"

	"github.com/globalsign/mgo/bson"
)

type AccountTx struct {
	Name   string `bson:"name"`
	Time   int64  `bson:"time"`
	TxHash string `bson:"txHash"`
}

type AccountPubkey struct {
	Name   string `bson:"name"`
	Pubkey string `bson:"pubkey"`
}

type Account struct {
	Name       string  `bson:"name"`
	CreateTime int64   `bson:"createTime"`
	Creator    string  `bson:"creator"`
	Balance    float64 `bson:"balance"`
	AccountPb  []byte  `bson:"accountPb"`
}

func GetAccountTxByName(name string, start, limit int) ([]*AccountTx, error) {
	accountTxC, err := GetCollection(CollectionAccountTx)
	if err != nil {
		return nil, err
	}
	//query := bson.M{
	//	"balance": bson.M{"$ne": 0},
	//}
	query := bson.M{
		"name": name,
	}
	var accountTxList []*AccountTx
	err = accountTxC.Find(query).Sort("-time").Skip(start).Limit(limit).All(&accountTxList)
	if err != nil {
		return nil, err
	}
	return accountTxList, nil
}

func GetAccountTxNumber(name string) (int, error) {
	accountTxC, err := GetCollection(CollectionAccountTx)
	if err != nil {
		return nil, err
	}
	return accountTxC.Find(bson.M{}).Count()
}

func GetAccountPubkeyByName(name string) ([]*AccountPubkey, error) {
	accountPubC, err := GetCollection(CollectionAccountPubkey)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"name": name,
	}
	var accountPubkeyList []*AccountPubkey
	err = accountPubC.Find(query).All(&accountPubkeyList)
	if err != nil {
		return nil, err
	}
	return accountPubkeyList, nil
}

func GetAccountPubkeyByPubkey(pubkey string) ([]*AccountPubkey, error) {
	accountPubC, err := GetCollection(CollectionAccountPubkey)
	if err != nil {
		return nil, err
	}
	query := bson.M{
		"pubkey": pubkey,
	}
	var accountPubkeyList []*AccountPubkey
	err = accountPubC.Find(query).All(&accountPubkeyList)
	if err != nil {
		return nil, err
	}
	return accountPubkeyList, nil
}

func GetAccounts(start, limit int) ([]*Account, error) {
	accountC, err := GetCollection(CollectionAccount)
	if err != nil {
		return nil, err
	}
	//query := bson.M{
	//	"balance": bson.M{"$ne": 0},
	//}
	query := bson.M{}
	var accountList []*Account
	err = accountC.Find(query).Sort("-balance").Skip(start).Limit(limit).All(&accountList)
	if err != nil {
		return nil, err
	}

	return accountList, nil
}

func GetAccountByName(name string) (*Account, error) {
	accountC, err := GetCollection(CollectionAccount)
	if err != nil {
		return nil, err
	}

	query := bson.M{
		"name": name,
	}
	var account *Account
	err = accountC.Find(query).One(&account)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func GetAccountsTotalLen() (int, error) {
	accountC, err := GetCollection(CollectionAccount)
	if err != nil {
		return 0, err
	}
	//query := bson.M{
	//	"balance": bson.M{"$ne": 0},
	//}
	query := bson.M{}
	return accountC.Find(query).Count()
}

func GetAccountLastPage(eachPage int64) (int64, error) {
	accountC, err := GetCollection(CollectionAccount)
	if err != nil {
		log.Println("GetAccounts get collection error:", err)
		return 0, err
	}

	query := bson.M{
		"balance": bson.M{"$ne": 0},
	}
	totalLen, _ := accountC.Find(query).Count()
	totalLenInt64 := int64(totalLen)

	var pageLast int64
	if totalLenInt64%eachPage == 0 {
		pageLast = totalLenInt64 / eachPage
	} else {
		pageLast = totalLenInt64/eachPage + 1
	}

	if pageLast == 0 {
		pageLast = 1
	}

	return pageLast, nil
}
