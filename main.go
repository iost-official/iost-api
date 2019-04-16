package main

import (
	"github.com/iost-official/iost-api/config"
	"github.com/iost-official/iost-api/controller"
	"github.com/iost-official/iost-api/middleware"
	"github.com/labstack/echo"
	echoMiddle "github.com/labstack/echo/middleware"
)

func main() {
	config.ReadConfig("")
	e := echo.New()
	e.Debug = true
	e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler
	e.Use(middleware.CorsHeader)
	e.Use(echoMiddle.Recover())
	e.Use(echoMiddle.Logger())

	// blocks
	/*  e.GET("/api/blocks", controller.GetBlocks) */
	// e.GET("/api/block/:id", controller.GetBlockDetail)

	// // transactions
	// e.GET("/api/txs", controller.GetTxs)
	// e.GET("/api/tx/:id", controller.GetTxnDetail)

	// // accounts
	// e.GET("/api/accounts", controller.GetAccounts)
	e.GET("/iost-api/accounts/:id", controller.GetAccountDetail)
	e.GET("/iost-api/pledge/:id", controller.GetAccountPledge)
	e.GET("/iost-api/accountTx", controller.GetAccountTxs)
	e.GET("/iost-api/contractTx", controller.GetContractTxs)
	e.GET("/iost-api/candidates", controller.GetCandidates)

	// search
	// e.GET("/api/search/:id", controller.GetSearch)

	// e.GET("/api/dropDatabase", controller.DropDatabase)

	e.Logger.Fatal(e.Start(":" + "8002"))
}
