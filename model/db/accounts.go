package db

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/iost-api/model/blockchain"
	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
)

type AccountTx struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name"`
	Time           int64         `bson:"time"`
	TxHash         string        `bson:"txHash"`
	TransferTokens []string      `bson:"tokens,omitempty"`
}

type Account struct {
	Name        string         `bson:"name" json:"name"`
	CreateTime  int64          `bson:"createTime" json:"create_time"`
	Creator     string         `bson:"creator" json:"creator"`
	AccountInfo *rpcpb.Account `bson:"accountInfo" json:"account_info"`
	Pubkeys     []string       `bson:"pubkeys" json:"-"`
}

type PledgeInfo struct {
	MyPledge    map[string]float64 `json:"my_pledge"`
	PledgeForMe map[string]float64 `json:"pledge_for_me"`
}

func NewAccount(name string, time int64, creator string) *Account {
	return &Account{
		Name:       name,
		CreateTime: time,
		Creator:    creator,
	}
}

func getAccTxQuery(name string, onlyTransfer bool, transferToken string) bson.M {
	query := bson.M{
		"name": name,
	}
	if onlyTransfer {
		if transferToken == "" {
			query["tokens"] = bson.M{
				"$nin": []interface{}{nil},
			}
		} else {
			query["tokens"] = transferToken
		}
	}
	return query
}

func GetAccountTxByName(name string, start, limit int, onlyTransfer bool, transferToken string) ([]*AccountTx, error) {
	accountTxC := GetCollection(CollectionAccountTx)

	query := getAccTxQuery(name, onlyTransfer, transferToken)
	var accountTxList []*AccountTx
	err := accountTxC.Find(query).Sort("-time").Skip(start).Limit(limit).All(&accountTxList)
	if err != nil {
		return nil, err
	}
	return accountTxList, nil
}

func GetAccountTxByNameAndPos(name, pos string, limit int, onlyTransfer bool, transferToken string) ([]*AccountTx, error) {
	accountTxC := GetCollection(CollectionAccountTx)

	query := getAccTxQuery(name, onlyTransfer, transferToken)
	query["_id"] = bson.M{"$lt": pos}
	var accountTxList []*AccountTx
	err := accountTxC.Find(query).Sort("-_id").Limit(limit).All(&accountTxList)
	if err != nil {
		return nil, err
	}
	return accountTxList, nil
}

func GetAccountTxNumber(name string, onlyTransfer bool, transferToken string) (int, error) {
	accountTxC := GetCollection(CollectionAccountTx)

	query := getAccTxQuery(name, onlyTransfer, transferToken)
	return accountTxC.Find(query).Count()
}

func GetAccounts(start, limit int) ([]*Account, error) {
	accountC := GetCollection(CollectionAccount)
	//query := bson.M{
	//	"balance": bson.M{"$ne": 0},
	//}
	query := bson.M{}
	var accountList []*Account
	err := accountC.Find(query).Sort("-balance").Skip(start).Limit(limit).All(&accountList)
	if err != nil {
		return nil, err
	}

	return accountList, nil
}

func GetAccountPledge(name string) (*PledgeInfo, error) {
	result := &PledgeInfo{
		MyPledge:    make(map[string]float64),
		PledgeForMe: make(map[string]float64),
	}
	accountC := GetCollection(CollectionAccount)
	query := bson.M{
		"name": name,
	}
	var accountList []*Account
	err := accountC.Find(query).All(&accountList)
	if err != nil {
		return nil, err
	}
	if len(accountList) == 0 {
		return nil, fmt.Errorf("account name %v not exist", name)
	}
	for _, item := range accountList[0].AccountInfo.GasInfo.PledgedInfo {
		result.MyPledge[item.Pledger] = item.Amount
	}
	query2 := bson.M{
		"accountInfo.gasinfo.pledgedinfo.pledger": name,
	}
	accountList = make([]*Account, 0)
	err = accountC.Find(query2).All(&accountList)
	if err != nil {
		return nil, err
	}
	for _, acc := range accountList {
		for _, item := range acc.AccountInfo.GasInfo.PledgedInfo {
			if item.Pledger == name {
				result.PledgeForMe[acc.Name] = item.Amount
			}
		}
	}
	return result, nil
}

func GetAccountByName(name string) (*Account, error) {
	accountC := GetCollection(CollectionAccount)

	query := bson.M{
		"name": name,
	}
	var account *Account
	err := accountC.Find(query).One(&account)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func GetAccountsByNames(names []string) ([]*Account, error) {
	accountC := GetCollection(CollectionAccount)
	query := bson.M{
		"name": bson.M{
			"$in": names,
		},
	}

	var accounts []*Account
	err := accountC.Find(query).All(&accounts)

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func GetAccountsByPubkey(pubkey string) ([]*Account, error) {
	accountC := GetCollection(CollectionAccount)
	query := bson.M{
		"pubkeys": pubkey,
	}
	var accounts []*Account
	err := accountC.Find(query).All(&accounts)
	if err != nil {
		return nil, err
	}
	return accounts, nil

}

func GetAccountsTotalLen() (int, error) {
	accountC := GetCollection(CollectionAccount)
	//query := bson.M{
	//	"balance": bson.M{"$ne": 0},
	//}
	query := bson.M{}
	return accountC.Find(query).Count()
}

func GetAccountLastPage(eachPage int64) (int64, error) {
	accountC := GetCollection(CollectionAccount)

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

func isContract(name string) bool {
	return strings.HasPrefix(name, "Contract") || strings.Index(name, ".") > -1
}

func getAccountsByRPC(accounts map[string]struct{}) []*rpcpb.Account {
	if len(accounts) == 0 {
		return nil
	}
	accCh := make(chan *rpcpb.Account, len(accounts))
	for name := range accounts {
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
	var ret []*rpcpb.Account
	for accountInfo := range accCh {
		i++
		if accountInfo != nil {
			ret = append(ret, accountInfo)
		}
		if i == len(accounts) {
			break
		}
	}
	return ret
}

func getPubkeys(acc *rpcpb.Account) []string {
	var ret []string
	pkSet := make(map[string]struct{})
	for _, perm := range acc.Permissions {
		for _, item := range perm.Items {
			if item.IsKeyPair {
				pkSet[item.Id] = struct{}{}
			}
		}
	}
	for _, group := range acc.Groups {
		for _, item := range group.Items {
			if item.IsKeyPair {
				pkSet[item.Id] = struct{}{}
			}
		}
	}
	for pk := range pkSet {
		ret = append(ret, pk)
	}
	sort.Strings(ret)
	return ret
}

func getContractsByRPC(contracts map[string]struct{}) []*rpcpb.Contract {
	if len(contracts) == 0 {
		return nil
	}
	contCh := make(chan *rpcpb.Contract, len(contracts))
	for id := range contracts {
		go func(id string) {
			contractInfo, err := blockchain.GetContract(id, false)
			if err != nil {
				contCh <- nil
			} else {
				contCh <- contractInfo
			}
		}(id)
	}

	var i int
	var ret []*rpcpb.Contract
	for contractInfo := range contCh {
		i++
		if contractInfo != nil {
			ret = append(ret, contractInfo)
		}
		if i == len(contracts) {
			break
		}
	}
	return ret
}

func retryWriteMgo(b *mgo.Bulk, wg *sync.WaitGroup) {
	defer wg.Done()

	var retryTime int
	for {
		if _, err := b.Run(); err != nil {
			log.Println("fail to write data to mongo ", err)
			time.Sleep(time.Second)
			retryTime++
			if retryTime > 10 {
				log.Fatalln("fail to write data to mongo, retry time exceeds")
			}
			continue
		}
		return
	}
}

func gatherAccountTxs(accountTxs map[string]*AccountTx, name, txHash string, time int64, token *string) {
	if isContract(name) {
		return
	}
	key := name + "@" + txHash

	accTx := accountTxs[key]
	if accTx == nil {
		accTx = &AccountTx{"", name, time, txHash, nil}
	}
	if token != nil {
		var exist bool
		for _, t := range accTx.TransferTokens {
			if t == *token {
				exist = true
				break
			}
		}

		if !exist {
			accTx.TransferTokens = append(accTx.TransferTokens, *token)
		}
	}
	accountTxs[key] = accTx
}

func ProcessTxsForAccount(txs []*rpcpb.Transaction, blockTime int64) {

	accTxC := GetCollection(CollectionAccountTx)
	accTxB := accTxC.Bulk()

	accountC := GetCollection(CollectionAccount)
	accountB := accountC.Bulk()

	contractC := GetCollection(CollectionContract)
	contractB := contractC.Bulk()

	contractTxC := GetCollection(CollectionContractTx)
	contractTxB := contractTxC.Bulk()

	updatedAccounts := make(map[string]struct{})
	updatePubkey := make(map[string]bool)
	updatedContracts := make(map[string]struct{})

	accountTxs := make(map[string]*AccountTx)

	for _, t := range txs {

		for _, r := range t.TxReceipt.Receipts {

			// transfer
			if r.FuncName == "token.iost/transfer" {
				var params []string
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) == 5 {
					gatherAccountTxs(accountTxs, params[1], t.Hash, blockTime, &params[0])
					gatherAccountTxs(accountTxs, params[2], t.Hash, blockTime, &params[0])

					if !isContract(params[1]) {
						updatedAccounts[params[1]] = struct{}{}
					}
					if !isContract(params[2]) {
						updatedAccounts[params[2]] = struct{}{}
					}
				}
			}

			// create user
			if r.FuncName == "auth.iost/signUp" {
				var params []string
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) == 3 {
					accountB.Upsert(bson.M{"name": params[0]}, bson.M{"$set": bson.M{"createTime": blockTime, "creator": t.Publisher}})

					gatherAccountTxs(accountTxs, params[0], t.Hash, blockTime, nil)
					updatePubkey[params[0]] = true
				}
			}

			// update pubkey
			if strings.HasPrefix(r.FuncName, "auth.iost/") {
				var params []string
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) > 0 {
					if !isContract(params[0]) {
						updatedAccounts[params[0]] = struct{}{}
						updatePubkey[params[0]] = true
					}
				}
			}
		}

		for _, a := range t.Actions {

			if a.Contract == "system.iost" && a.ActionName == "initSetCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 2 {
					contractB.Insert(NewContract(params[0], blockTime, t.Publisher))
					contractTxB.Insert(&ContractTx{params[0], blockTime, t.Hash})

					updatedContracts[params[0]] = struct{}{}
				}
			}

			if a.Contract == "system.iost" && a.ActionName == "setCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {

				contractID := "Contract" + t.Hash
				contractB.Insert(NewContract(contractID, blockTime, t.Publisher))
				contractTxB.Insert(&ContractTx{contractID, blockTime, t.Hash})

				updatedContracts[contractID] = struct{}{}
			}

			if a.Contract == "gas.iost" && (a.ActionName == "pledge" || a.ActionName == "unpledge") &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 3 {
					if !isContract(params[0]) {
						updatedAccounts[params[0]] = struct{}{}
					}
				}
			}

			contractTxB.Insert(&ContractTx{a.Contract, blockTime, t.Hash})
		}

		if t.Publisher != "_Block_Base" {
			gatherAccountTxs(accountTxs, t.Publisher, t.Hash, blockTime, nil)
		}

	}

	for _, accTx := range accountTxs {
		accTxB.Insert(accTx)
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		accountInfos := getAccountsByRPC(updatedAccounts)
		for _, acc := range accountInfos {
			accountB.Update(bson.M{"name": acc.Name}, bson.M{"$set": bson.M{"accountInfo": acc}})
			if updatePubkey[acc.Name] {
				pks := getPubkeys(acc)
				if len(pks) > 0 {
					accountB.Update(bson.M{"name": acc.Name}, bson.M{"$set": bson.M{"pubkeys": pks}})
				}
			}
		}
		wg.Done()
	}()

	go func() {
		contractInfos := getContractsByRPC(updatedContracts)
		for _, cont := range contractInfos {
			contractB.Update(bson.M{"id": cont.Id}, bson.M{"$set": bson.M{"contractInfo": cont}})
		}
		wg.Done()
	}()
	wg.Wait()

	wg.Add(4)
	go retryWriteMgo(accTxB, wg)
	go retryWriteMgo(accountB, wg)
	go retryWriteMgo(contractB, wg)
	go retryWriteMgo(contractTxB, wg)
	wg.Wait()
}
