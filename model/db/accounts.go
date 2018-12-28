package db

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/iost-api/model/blockchain"
	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
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
	Name        string         `bson:"name"`
	CreateTime  int64          `bson:"createTime"`
	Creator     string         `bson:"creator"`
	Balance     float64        `bson:"balance"`
	AccountInfo *rpcpb.Account `bson:"accountInfo"`
	// AccountPb   []byte         `bson:"accountPb"`
}

func NewAccount(name string, time int64, creator string) *Account {
	return &Account{
		Name:       name,
		CreateTime: time,
		Creator:    creator,
	}
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
		return 0, err
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

func printError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ProcessTxsForAccount(txs []*rpcpb.Transaction, blockTime int64) {

	accTxC, err := GetCollection(CollectionAccountTx)
	printError(err)
	accTxB := accTxC.Bulk()

	accountPubC, err := GetCollection(CollectionAccountPubkey)
	printError(err)
	accountPubB := accountPubC.Bulk()

	accountC, err := GetCollection(CollectionAccount)
	printError(err)
	accountB := accountC.Bulk()

	contractC, err := GetCollection(CollectionContract)
	printError(err)
	contractB := contractC.Bulk()

	contractTxC, err := GetCollection(CollectionContractTx)
	printError(err)
	contractTxB := contractTxC.Bulk()

	updatedAccounts := make(map[string]struct{})

	for _, t := range txs {

		for _, a := range t.Actions {

			// create account
			if a.Contract == "auth.iost" && a.ActionName == "SignUp" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 3 {
					account := NewAccount(params[0], blockTime, t.Publisher)
					accountB.Insert(account)

					accountPubB.Insert(&AccountPubkey{params[0], params[1]})
					if params[1] != params[2] {
						accountPubB.Insert(&AccountPubkey{params[0], params[2]})
					}

					accTxB.Insert(&AccountTx{params[0], blockTime, t.Hash})
				}
			}

			if a.Contract == "system.iost" && a.ActionName == "InitSetCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 2 {
					contractB.Insert(NewContract(params[0], blockTime, t.Publisher))
					contractTxB.Insert(&ContractTx{params[0], blockTime, t.Hash})
				}
			}

			if a.Contract == "system.iost" && a.ActionName == "SetCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				contractB.Insert(NewContract("Contract"+t.Hash, blockTime, t.Publisher))
				contractTxB.Insert(&ContractTx{"Contract" + t.Hash, blockTime, t.Hash})
			}

			contractTxB.Insert(&ContractTx{a.Contract, blockTime, t.Hash})

		}

		for _, r := range t.TxReceipt.Receipts {

			if r.FuncName == "token.iost/transfer" {
				var params []string
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) == 5 {
					accTxB.Insert(&AccountTx{params[1], blockTime, t.Hash})
					accTxB.Insert(&AccountTx{params[2], blockTime, t.Hash})

					if strings.Index(params[1], ".") == -1 {
						updatedAccounts[params[1]] = struct{}{}
					}
					if strings.Index(params[2], ".") == -1 {
						updatedAccounts[params[2]] = struct{}{}
					}
				}
			}
		}
	}

	if len(updatedAccounts) > 0 {
		accCh := make(chan *rpcpb.Account, len(updatedAccounts))
		for name, _ := range updatedAccounts {
			go func(name string) {
				accountInfo, err := blockchain.GetAccount(name, false)
				if err != nil {
					accCh <- nil
				} else {
					accCh <- accountInfo
				}
			}(name)
		}

		var i int
		for accountInfo := range accCh {
			i++
			if accountInfo != nil {
				accountB.Update(bson.M{"name": accountInfo.Name}, bson.M{"accountInfo": accountInfo})
			}
			if i == len(updatedAccounts) {
				break
			}
		}

	}

	_, err = accTxB.Run()
	printError(err)
	_, err = accountPubB.Run()
	printError(err)
	_, err = accountB.Run()
	printError(err)
	contractB.Run()
	contractTxB.Run()
}
