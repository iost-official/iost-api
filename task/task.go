package main

import (
	"sync"

	"github.com/iost-official/iost-api/config"
	"github.com/iost-official/iost-api/task/cron"
)

var ws = new(sync.WaitGroup)

func main() {
	config.ReadConfig()

	// start tasks
	ws.Add(7)
	go cron.UpdateBlocks(ws)
	go cron.ProcessFailedSyncBlocks(ws)
	go cron.UpdateTxns(ws, 0)
	go cron.UpdateTxns(ws, 1)
	go cron.UpdateRpcErrTxns(ws)
	go cron.UpdateBlockPay(ws)
	go cron.UpdateAccounts(ws)
	ws.Wait()
}
