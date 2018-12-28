package db

var (
	MongoLink = "mongodb://47.244.109.92:27017"
	Db        string
)

const (
	CollectionBlocks        = "blocks"
	CollectionTmpTxs        = "tmpTxs"
	CollectionTxs           = "txs"
	CollectionRpcTxs        = "rpcErrTxs"
	CollectionFlatTx        = "flatxs"
	CollectionFBlocks       = "fBlocks"
	CollectionAccount       = "accounts"
	CollectionAccountTx     = "accountTx"
	CollectionAccountPubkey = "accountPubkey"
	CollectionContract      = "contracts"
	CollectionContractTx    = "contractTx"
	CollectionTaskCursor    = "taskCursors"
	CollectionBlockPay      = "blockPays"
	CollectionApplyIost     = "applyTestIOST"
)
