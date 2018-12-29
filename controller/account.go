package controller

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"net/http"
	"strconv"
	"strings"

	"log"
	"sync"

	"github.com/iost-official/iost-api/model/db"
	"github.com/labstack/echo"

	"github.com/btcsuite/btcutil/base58"
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
	Account string        `json:"account"`
	TxnList []*db.TxStore `json:"txnList"`
	TxnLen  int           `json:"txnLen"`
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

func parity(bit []byte) []byte {
	crc32q := crc32.MakeTable(crc32.Koopman)
	crc := crc32.Checksum(bit, crc32q)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, crc)
	return bs
}

func getIDByPubkey(pubkey string) string {
	pbk := base58.Decode(pubkey)
	return "IOST" + base58.Encode(append(pbk, parity(pbk)...))
}

func GetAccountDetail(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return errors.New("id or pubkey required")
	}
	names := []string{id}

	if len(id) > 11 { // not an account name
		if !strings.HasPrefix(id, "IOST") {
			id = getIDByPubkey(id)
		}

		accountPubkeys, err := db.GetAccountPubkeyByPubkey(id)
		if err != nil {
			return err
		}
		names = make([]string, len(accountPubkeys))
		for i, ap := range accountPubkeys {
			names[i] = ap.Name
		}

	}
	accounts, err := db.GetAccountsByNames(names)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FormatResponse(struct {
		Accounts []*db.Account `json:"accounts"`
		Count    int           `json:"count"`
	}{
		accounts,
		len(accounts),
	}))
}

func GetAccountTxs(c echo.Context) error {
	account := c.QueryParam("account")
	if account == "" {
		return errors.New("address requied")
	}

	page := c.QueryParam("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		pageInt = 1
	}

	start := (pageInt - 1) * AccountTxEachPage
	txHashes, err := db.GetAccountTxByName(account, start, AccountTxEachPage)
	if err != nil {
		return err
	}
	hashes := make([]string, len(txHashes))
	for i, t := range txHashes {
		hashes[i] = t.TxHash
	}

	output := &AccountTxsOutput{
		Account: account,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// get tx detail
	go func() {
		defer wg.Done()
		txs, err := db.GetTxsByHash(hashes)
		if err != nil {
			log.Println("GetTxsByHash error: ", err)
			return
		}
		output.TxnList = txs
	}()
	// get account len
	go func() {
		defer wg.Done()
		totalLen, err := db.GetAccountTxNumber(account)
		if err != nil {
			log.Println("GetAccountTxNumber error:", err)
			return
		}
		output.TxnLen = totalLen
	}()
	wg.Wait()

	return c.JSON(http.StatusOK, FormatResponse(output))
}
