package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/globalsign/mgo/bson"
	"github.com/iost-official/iost-api/model/blockchain"
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
	Address  string           `json:"address"`
	TxnList  []*db.JsonFlatTx `json:"txnList"`
	TxnLen   int              `json:"txnLen"`
	PageLast int              `json:"pageLast"`
}

func init() {
	gcapHttpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns: 10,
		},
	}
}

func calLastPage(total int) int {
	var lastPage int
	if total%AccountEachPage == 0 {
		lastPage = total / AccountEachPage
	} else {
		lastPage = total/AccountEachPage + 1
	}

	if lastPage > AccountMaxPage { // ?
		lastPage = AccountMaxPage
	}
	return lastPage
}

func GetAccounts(c echo.Context) error {
	page := c.QueryParam("p")

	pageInt, _ := strconv.Atoi(page)
	if pageInt <= 0 {
		pageInt = 1
	}

	start := (pageInt - 1) * AccountEachPage
	accountList, err := db.GetAccounts(start, AccountEachPage)
	if err != nil {
		return err
	}

	accountTotalLen, err := db.GetAccountsTotalLen()
	if err != nil {
		return err
	}
	lastPage := calLastPage(accountTotalLen)

	output := &AccountsOutput{
		AccountList: accountList,
		Page:        pageInt,
		PagePrev:    pageInt - 1,
		PageNext:    pageInt + 1,
		PageLast:    lastPage,
		TotalLen:    accountTotalLen,
	}

	return c.JSON(http.StatusOK, FormatResponse(output))
}

func GetAccountDetail(c echo.Context) error {
	// TODO 检查地址格式
	address := c.Param("id")
	if address == "" {
		return errors.New("Address required")
	}

	col := db.GetCollection(db.CollectionAccount)

	if !(address[0:4] != "IOST" || address[0:8] != "Contract") {
		return errors.New("Invalid address")
	}

	account, err := db.GetAccountByAddress(address)
	// 如果记录不存在创建记录
	if err != nil && err.Error() == "not found" {
		account = &db.Account{
			address,
			0,
			0,
			0,
		}
		err = col.Insert(*account)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	toUpdate := bson.M{}
	txCount, err := db.GetAccountTxCount(address)
	if err == nil {
		account.TxCount = txCount
		toUpdate["tx_count"] = txCount
	}
	if address[0:4] == "IOST" { // IOST 地址获取余额
		balance, err := blockchain.GetBalance(address)
		if err == nil {
			account.Balance = float64(balance)
			toUpdate["balance"] = balance
		}
	}
	err = col.Update(bson.M{"address": address}, bson.M{"$set": toUpdate})

	if err != nil {
		return err
	}

	// 合约地址， 获取合约代码
	code := ""
	if address[0:8] == "Contract" {
		txhash := address[8:]
		txDetail, _ := db.GetFlatTxnDetailByHash(txhash)
		code = txDetail.Action.Data
	}

	return c.JSON(http.StatusOK, FormatResponse(struct {
		db.Account
		Code string `json:"code"`
	}{
		*account,
		code,
	}))
}

func GetAccountTxs(c echo.Context) error {
	address := c.Param("id")
	if address == "" {
		return errors.New("address requied")
	}

	page := c.QueryParam("p")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		pageInt = 1
	}

	start := (pageInt - 1) * AccountEachPage
	txnList, err := db.GetTxnListByAccount(address, start, AccountEachPage)
	if err != nil {
		return err
	}

	totalLen, _ := db.GetFlatTxnLenByAccount(address)
	pageLast := calLastPage(totalLen)

	output := &AccountTxsOutput{
		address,
		txnList,
		totalLen,
		pageLast,
	}

	return c.JSON(http.StatusOK, FormatResponse(output))
}
