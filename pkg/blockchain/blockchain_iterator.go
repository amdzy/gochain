package blockchain

type BlockchainIterator struct {
	currentHash []byte
	db          *DB
}

func (i *BlockchainIterator) Next() (*Block, error) {
	block, err := i.db.GetBlock(i.currentHash)
	if err != nil {
		return nil, err
	}

	i.currentHash = block.PrevBlockHash

	return block, nil
}
