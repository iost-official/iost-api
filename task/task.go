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
	ws.Add(6)
	// download block
	go cron.UpdateBlocks(ws)
	go cron.UpdateTxns(ws)
	go cron.UpdateAccounts(ws)
	ws.Wait()
}
