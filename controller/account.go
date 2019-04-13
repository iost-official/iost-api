package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"log"
	"sync"

	"github.com/iost-official/iost-api/model/db"
	"github.com/labstack/echo"
)

type AccountsOutput struct {
	AccountList []*db.Account `json:"accountList"`
	Page        int           `json:"page"`
	PagePrev    int           `json:"pagePrev"`
	PageNext    int           `json:"pageNext"`
	PageLast    int           `json:"pageLast"`
	TotalLen    int           `json:"totalLen"`
}

type AccountTxsOutput struct {
	Account string       `json:"account"`
	TxnList []*TxsOutput `json:"txnList"`
	TxnLen  int          `json:"txnLen"`
}

type ContractTxsOutput struct {
	Contract string       `json:"contract"`
	TxnList  []*TxsOutput `json:"txnList"`
	TxnLen   int          `json:"txnLen"`
}

//func GetAccounts(c echo.Context) error {
//	page := c.QueryParam("p")
//
//	pageInt, _ := strconv.Atoi(page)
//	if pageInt <= 0 {
//		pageInt = 1
//	}
//
//	start := (pageInt - 1) * AccountEachPage
//	accountList, err := db.GetAccounts(start, AccountEachPage)
//	if err != nil {
//		return err
//	}
//
//	accountTotalLen, err := db.GetAccountsTotalLen()
//	if err != nil {
//		return err
//	}
//	lastPage := calLastPage(accountTotalLen)
//
//	output := &AccountsOutput{
//		AccountList: accountList,
//		Page:        pageInt,
//		PagePrev:    pageInt - 1,
//		PageNext:    pageInt + 1,
//		PageLast:    lastPage,
//		TotalLen:    accountTotalLen,
//	}
//
//	return c.JSON(http.StatusOK, FormatResponse(output))
//}

func GetAccountPledge(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errors.New("id required")
	}
	pledge, err := db.GetAccountPledge(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, FormatResponse(pledge))
}

func GetAccountDetail(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errors.New("id or pubkey required")
	}

	var accounts []*db.Account
	var err error
	if len(id) > 11 { // not an account name

		accounts, err = db.GetAccountsByPubkey(id)
		if err != nil {
			return err
		}
		if len(accounts) == 0 {
			return errors.New("not found")
		}
	} else {
		account, err := db.GetAccountByName(id)
		if err != nil {
			return err
		}
		accounts = append(accounts, account)
	}

	return c.JSON(http.StatusOK, FormatResponse(struct {
		Accounts []*db.Account `json:"accounts"`
		Count    int           `json:"count"`
	}{
		accounts,
		len(accounts),
	}))
}

func GetContractTxs(c echo.Context) error {
	contractID := c.QueryParam("contractId")
	if contractID == "" {
		return errors.New("account requied")
	}

	ascending := c.QueryParam("ascending") == "1"

	var contractTxs []*db.ContractTx

	s := time.Now().UnixNano()
	pos := c.QueryParam("pos")
	if pos != "" {
		offset := c.QueryParam("offset")
		offsetInt, err := strconv.Atoi(offset)
		if err != nil || offsetInt <= 0 {
			offsetInt = ContractTxEachPage
		}
		if offsetInt > ContractTxMaxCount {
			offsetInt = ContractTxMaxCount
		}
		contractTxs, err = db.GetContractTxByIDAndPos(contractID, pos, offsetInt, ascending)
		if err != nil {
			return err
		}
	} else {
		page := c.QueryParam("page")
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt <= 0 {
			pageInt = 1
		}
		if pageInt > ContractTxMaxCount {
			pageInt = ContractTxMaxCount
		}
		start := (pageInt - 1) * ContractTxEachPage
		contractTxs, err = db.GetContractTxByID(contractID, start, ContractTxEachPage, ascending)
		if err != nil {
			return err
		}
	}
	s1 := time.Now().UnixNano()
	log.Printf("GetContractTx costs %d ns, contractId=%v", s1-s, contractID)

	hashes := make([]string, len(contractTxs))
	hashToUID := make(map[string]string)
	for i, t := range contractTxs {
		hashes[i] = t.TxHash
		hashToUID[t.TxHash] = t.BID.Hex()
	}

	output := &ContractTxsOutput{
		Contract: contractID,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// get tx detail
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("GetTxsByHash panic. err=%v", e)
			}
		}()
		s := time.Now().UnixNano()
		defer wg.Done()
		txs, err := db.GetTxsByHash(hashes)
		if err != nil {
			log.Println("GetTxsByHash error: ", err)
			return
		}
		for _, t := range txs {
			output.TxnList = append(output.TxnList, NewTxsOutputFromTxStore(t, hashToUID[t.Tx.Hash]))
		}
		log.Printf("GetTxDetail costs %d ns, contractId: %v", time.Now().UnixNano()-s, contractID)
	}()
	// get account len
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("GetAccountTxNumber panic. err=%v", e)
			}
		}()
		s := time.Now().UnixNano()
		defer wg.Done()
		totalLen, err := db.GetContractTxNumber(contractID)
		if err != nil {
			log.Println("GetAccountTxNumber error:", err)
			return
		}
		output.TxnLen = totalLen
		log.Printf("GetTxCount costs %d ns, contractId: %v", time.Now().UnixNano()-s, contractID)
	}()
	wg.Wait()
	log.Printf("GetTxDetailAndCount costs %d ns, contractId: %v", time.Now().UnixNano()-s1, contractID)

	return c.JSON(http.StatusOK, FormatResponse(output))
}

func GetAccountTxs(c echo.Context) error {
	account := c.QueryParam("account")
	if account == "" {
		return errors.New("account requied")
	}

	onlyTransfer := c.QueryParam("transfer") == "1"
	tokenName := c.QueryParam("token")
	ascending := c.QueryParam("ascending") == "1"

	var accountTxs []*db.AccountTx

	pos := c.QueryParam("pos")
	if pos != "" {
		offset := c.QueryParam("offset")
		offsetInt, err := strconv.Atoi(offset)
		if err != nil || offsetInt <= 0 {
			offsetInt = AccountTxEachPage
		}
		if offsetInt > AccountTxMaxCount {
			offsetInt = AccountTxMaxCount
		}
		accountTxs, err = db.GetAccountTxByNameAndPos(account, pos, offsetInt, onlyTransfer, tokenName, ascending)
		if err != nil {
			return err
		}
	} else {
		page := c.QueryParam("page")
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt <= 0 {
			pageInt = 1
		}
		if pageInt > AccountTxMaxCount {
			pageInt = AccountTxMaxCount
		}
		start := (pageInt - 1) * AccountTxEachPage
		accountTxs, err = db.GetAccountTxByName(account, start, AccountTxEachPage, onlyTransfer, tokenName, ascending)
		if err != nil {
			return err
		}
	}

	hashes := make([]string, len(accountTxs))
	hashToUID := make(map[string]string)
	for i, t := range accountTxs {
		hashes[i] = t.TxHash
		hashToUID[t.TxHash] = t.ID.Hex()
	}

	output := &AccountTxsOutput{
		Account: account,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// get tx detail
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("GetTxsByHash panic. err=%v", e)
			}
		}()
		defer wg.Done()
		txs, err := db.GetTxsByHash(hashes)
		if err != nil {
			log.Println("GetTxsByHash error: ", err)
			return
		}
		for _, t := range txs {
			output.TxnList = append(output.TxnList, NewTxsOutputFromTxStore(t, hashToUID[t.Tx.Hash]))
		}
	}()
	// get account len
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("GetAccountTxNumber panic. err=%v", e)
			}
		}()
		defer wg.Done()
		totalLen, err := db.GetAccountTxNumber(account, onlyTransfer, tokenName)
		if err != nil {
			log.Println("GetAccountTxNumber error:", err)
			return
		}
		output.TxnLen = totalLen
	}()
	wg.Wait()

	return c.JSON(http.StatusOK, FormatResponse(output))
}
