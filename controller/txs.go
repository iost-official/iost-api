package controller

import (
	"encoding/json"
	"log"

	"github.com/iost-official/iost-api/model/blockchain/rpcpb"
	"github.com/iost-official/iost-api/model/db"
)

const (
	TxEachPageNum = 25
	TxMaxPage     = 20
)

type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Token  string `json:"token"`
	Memo   string `json:"memo"`
}

func parseContractDataToTransfer(data string) *Transfer {
	var params []string
	err := json.Unmarshal([]byte(data), &params)
	if err != nil || len(params) != 5 {
		log.Println("unmarshal transfer params failed. err: ", err)
		return nil
	}
	return &Transfer{
		From:   params[1],
		To:     params[2],
		Amount: params[3],
		Token:  params[0],
		Memo:   params[4],
	}
}

type TxsOutput struct {
	*db.TxStore
	Transfers []*Transfer `json:"transfers"`
	UniqID    string      `json:"uniq_id"`
}

func NewTxsOutputFromTxStore(tx *db.TxStore, uniqID string) *TxsOutput {
	ret := &TxsOutput{TxStore: tx, UniqID: uniqID}
	for _, receipt := range tx.Tx.TxReceipt.Receipts {
		if receipt.FuncName == "token.iost/transfer" &&
			tx.Tx.TxReceipt.StatusCode == rpcpb.TxReceipt_SUCCESS {
			trans := parseContractDataToTransfer(receipt.Content)
			if trans != nil {
				ret.Transfers = append(ret.Transfers, trans)
			}
		}
	}
	return ret
}

/* func GetTxnDetail(c echo.Context) error { */
// txHash := c.Param("id")

// if txHash == "" {
// return nil
// }

// txnOut, err := model.GetDetailTxn(txHash)

// if err != nil {
// return err
// }

// return c.JSON(http.StatusOK, FormatResponse(txnOut))
// }

// func GetTxs(c echo.Context) error {
// page := c.QueryParam("page")
// address := c.QueryParam("account")
// blk := c.QueryParam("block")

// pageInt64, err := strconv.ParseInt(page, 10, 64)

// if err != nil || pageInt64 <= 0 {
// pageInt64 = 1
// }

// blockInt64, err := strconv.ParseInt(blk, 10, 64)

// if err != nil {
// blockInt64 = -1
// }

// txList, err := model.GetFlatTxnSlicePage(pageInt64, TxEachPageNum, blockInt64, address)

// if err != nil {
// return err
// }

// var (
// lastPage int64
// totalLen int
// )

// if address != "" {
// // get total page count for specific account
// lastPage, _ = db.GetFlatTxPageCntWithAddress(TxEachPageNum, address)
// totalLen, _ = db.GetTotalFlatTxnLen(address, -1)
// } else if blk != "" {
// // get total page count for specific block
// lastPage, _ = db.GetFlatTxPageCntWithBlk(TxEachPageNum, blockInt64)
// totalLen, _ = db.GetTotalFlatTxnLen("", blockInt64)
// } else {
// // get all page count for all transactions
// lastPage, _ = db.GetFlatTxTotalPageCnt(TxEachPageNum, "", -1)
// totalLen, _ = db.GetTotalFlatTxnLen("", -1)
// }

// if lastPage > TxMaxPage {
// lastPage = TxMaxPage
// }

// output := &TxsOutput{
// TxList:   txList,
// Page:     pageInt64,
// PagePrev: pageInt64 - 1,
// PageNext: pageInt64 + 1,
// PageLast: lastPage,
// TotalLen: totalLen,
// }

// return c.JSON(http.StatusOK, FormatResponse(output))
/* } */
