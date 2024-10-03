package main

type Blockchain struct {
	tip []byte
	db  *DB
}

func (bc *Blockchain) AddBlock(data string) error {
	tip, err := bc.db.GetTip()
	if err != nil {
		return err
	}

	block := NewBlock(data, tip)
	err = bc.db.AddBlock(block)
	if err != nil {
		return err
	}

	bc.tip = block.Hash

	return nil
}

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

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func NewBlockchain(db *DB) (*Blockchain, error) {
	tip, err := db.GetTip()
	if err != nil {
		return nil, err
	}
	return &Blockchain{db: db, tip: tip}, nil
}
