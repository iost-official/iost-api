package db

import (
	"encoding/hex"
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

func GetAccountTxByName(name string, start, limit int, onlyTransfer bool, transferToken string, ascending bool) ([]*AccountTx, error) {
	accountTxC := GetCollection(CollectionAccountTx)

	query := getAccTxQuery(name, onlyTransfer, transferToken)
	var accountTxList []*AccountTx
	var sort = "-time"
	if ascending {
		sort = "time"
	}
	err := accountTxC.Find(query).Sort(sort).Skip(start).Limit(limit).All(&accountTxList)
	if err != nil {
		return nil, err
	}
	return accountTxList, nil
}

func GetAccountTxByNameAndPos(name, pos string, limit int, onlyTransfer bool, transferToken string, ascending bool) ([]*AccountTx, error) {
	d, err := hex.DecodeString(pos)
	if err != nil {
		return nil, err
	}
	accountTxC := GetCollection(CollectionAccountTx)

	query := getAccTxQuery(name, onlyTransfer, transferToken)
	query["_id"] = bson.M{"$lt": bson.ObjectId(d)}
	var accountTxList []*AccountTx
	var sort = "-_id"
	if ascending {
		sort = "_id"
		query["_id"] = bson.M{"$gt": bson.ObjectId(d)}
	}
	err = accountTxC.Find(query).Sort(sort).Limit(limit).All(&accountTxList)
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
	err := accountC.Find(query).Limit(100).All(&accounts)
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

func getCandidatesByRPC(candidates map[string]struct{}) map[string]*rpcpb.GetProducerVoteInfoResponse {
	if len(candidates) == 0 {
		return nil
	}

	var m sync.Mutex
	var wg sync.WaitGroup
	ret := make(map[string]*rpcpb.GetProducerVoteInfoResponse, len(candidates))

	wg.Add(len(candidates))
	for id := range candidates {
		go func(id string) {
			producer, err := blockchain.GetProducer(id, false)
			if err == nil {
				m.Lock()
				ret[id] = producer
				m.Unlock()
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
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

func gatherContractTxs(contractTxs map[string]*ContractTx, cid, txHash string, time int64) {
	if !isContract(cid) {
		return
	}
	key := cid + "@" + txHash

	contractTxs[key] = &ContractTx{
		ID:     cid,
		Time:   time,
		TxHash: txHash,
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

	candidateC := GetCollection(CollectionCandidate)
	candidateB := candidateC.Bulk()

	updatedAccounts := make(map[string]struct{})
	updatePubkey := make(map[string]bool)
	updatedContracts := make(map[string]struct{})
	updatedCandidates := make(map[string]struct{})

	accountTxs := make(map[string]*AccountTx)
	contractTxs := make(map[string]*ContractTx)

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
			if r.FuncName == "auth.iost/signUp" && t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) == 3 {
					accountB.Upsert(bson.M{"name": params[0]}, bson.M{"$set": bson.M{"createTime": blockTime, "creator": t.Publisher}})

					gatherAccountTxs(accountTxs, params[0], t.Hash, blockTime, nil)
					updatePubkey[params[0]] = true
				}
			}

			// update pubkey
			if strings.HasPrefix(r.FuncName, "auth.iost/") && t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []interface{}
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) > 0 {
					name := params[0].(string)
					if !isContract(name) {
						updatedAccounts[name] = struct{}{}
						updatePubkey[name] = true
					}
				}
			}

			// update candidate
			if strings.HasPrefix(r.FuncName, "vote_producer.iost/") && t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []interface{}
				err := json.Unmarshal([]byte(r.Content), &params)
				if err == nil && len(params) > 0 {
					var candidate string

					abi := strings.Split(r.FuncName, "/")[1]
					switch abi {
					case "applyRegister", "applyUnregister", "approveRegister", "approveUnregister", "forceUnregister",
						"unregister", "updateProducer", "logInProducer", "logOutProducer":
						candidate = params[0].(string)
					case "vote", "unvote":
						candidate = params[1].(string)
					case "voteFor":
						candidate = params[2].(string)
					}
					updatedCandidates[candidate] = struct{}{}
				}
			}

			// update contract tx
			gatherContractTxs(contractTxs, r.FuncName[:strings.Index(r.FuncName, "/")], t.Hash, blockTime)
		}

		for _, a := range t.Actions {

			if a.Contract == "system.iost" && a.ActionName == "initSetCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 2 {
					contractB.Upsert(bson.M{"id": params[0]}, bson.M{"$set": bson.M{"createTime": blockTime, "creator": t.Publisher}})
					gatherContractTxs(contractTxs, params[0], t.Hash, blockTime)

					updatedContracts[params[0]] = struct{}{}
				}
			}

			if a.Contract == "system.iost" && a.ActionName == "setCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {

				contractID := "Contract" + t.Hash
				contractB.Upsert(bson.M{"id": contractID}, bson.M{"$set": bson.M{"createTime": blockTime, "creator": t.Publisher}})
				gatherContractTxs(contractTxs, contractID, t.Hash, blockTime)

				updatedContracts[contractID] = struct{}{}
			}

			if a.Contract == "system.iost" && a.ActionName == "updateCode" &&
				t.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
				var params []string
				err := json.Unmarshal([]byte(a.Data), &params)
				if err == nil && len(params) == 3 {
					gatherContractTxs(contractTxs, params[0], t.Hash, blockTime)
					updatedContracts[params[0]] = struct{}{}
				}
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

			gatherContractTxs(contractTxs, a.Contract, t.Hash, blockTime)
		}

		if t.Publisher != "base.iost" {
			gatherAccountTxs(accountTxs, t.Publisher, t.Hash, blockTime, nil)
		}

	}

	for _, accTx := range accountTxs {
		accTxB.Insert(accTx)
	}
	for _, contTx := range contractTxs {
		contractTxB.Insert(contTx)
	}

	wg := new(sync.WaitGroup)
	wg.Add(3)
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

	go func() {
		candidateInfos := getCandidatesByRPC(updatedCandidates)
		for name, cand := range candidateInfos {
			candidateB.Upsert(bson.M{"name": name}, bson.M{"$set": bson.M{"candidateInfo": cand}})
		}
		wg.Done()
	}()
	wg.Wait()

	wg.Add(5)
	go retryWriteMgo(accTxB, wg)
	go retryWriteMgo(accountB, wg)
	go retryWriteMgo(contractB, wg)
	go retryWriteMgo(contractTxB, wg)
	go retryWriteMgo(candidateB, wg)
	wg.Wait()
}
