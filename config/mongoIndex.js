db.txs.ensureIndex({"tx.hash":1},{ unique: true })
db.txs.ensureIndex({"blocknumber":1},{ unique: true })
db.blocks.ensureIndex({"number":1},{ unique: true })
db.blocks.ensureIndex({"hash":1},{ unique: true })
