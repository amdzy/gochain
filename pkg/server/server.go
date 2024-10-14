package server

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/transactions"
	"amdzy/gochain/pkg/utxo"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"slices"

	"github.com/vmihailenco/msgpack/v5"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var KnownNodes = []string{"localhost:6000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]transactions.Transaction)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getBlocks struct {
	AddrFrom string
}

type getData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return string(command)
}

func requestBlocks() error {
	for _, node := range KnownNodes {
		err := sendGetBlocks(node)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendAddr(address string) error {
	fmt.Println("Sending Addr")
	nodes := addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload, err := msgpack.Marshal(nodes)
	if err != nil {
		return err
	}
	request := append(commandToBytes("addr"), payload...)

	return sendData(address, request)
}

func sendBlock(addr string, b *blockchain.Block) error {
	blockSerialized, err := b.Serialize()
	if err != nil {
		return err
	}

	data := block{nodeAddress, blockSerialized}
	payload, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}

	request := append(commandToBytes("block"), payload...)

	return sendData(addr, request)
}

func sendData(addr string, data []byte) error {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return nil
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	return err
}

func sendInv(address, kind string, items [][]byte) error {
	inventory := inv{nodeAddress, kind, items}
	payload, err := msgpack.Marshal(inventory)
	if err != nil {
		return err
	}
	request := append(commandToBytes("inv"), payload...)

	return sendData(address, request)
}

func sendGetBlocks(address string) error {
	payload, err := msgpack.Marshal(getBlocks{nodeAddress})
	if err != nil {
		return err
	}

	request := append(commandToBytes("getblocks"), payload...)

	return sendData(address, request)
}

func sendGetData(address, kind string, id []byte) error {
	payload, err := msgpack.Marshal(getData{nodeAddress, kind, id})
	if err != nil {
		return err
	}
	request := append(commandToBytes("getdata"), payload...)

	return sendData(address, request)
}

func SendTx(addr string, tnx *transactions.Transaction) error {
	tnxSerialized, err := tnx.Serialize()
	if err != nil {
		return err
	}

	data := tx{nodeAddress, tnxSerialized}
	payload, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}

	request := append(commandToBytes("tx"), payload...)

	return sendData(addr, request)
}

func sendVersion(addr string, bc *blockchain.Blockchain) error {
	bestHeight, err := bc.GetBestHeight()
	fmt.Println(bestHeight)
	if err != nil {
		return err
	}

	payload, err := msgpack.Marshal(verzion{nodeVersion, bestHeight, nodeAddress})
	if err != nil {
		return err
	}

	request := append(commandToBytes("version"), payload...)

	return sendData(addr, request)
}

func handleAddr(request []byte) error {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	slices.Sort(KnownNodes)
	KnownNodes = slices.Compact(KnownNodes)
	fmt.Printf("There are %d known nodes now!\n", len(KnownNodes))
	fmt.Println(KnownNodes)
	return requestBlocks()
}

func handleBlock(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	blockData := payload.Block
	block, err := blockchain.DeserializeBlock(blockData)
	if err != nil {
		return err
	}

	fmt.Println("Received a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		err := sendGetData(payload.AddrFrom, "block", blockHash)
		if err != nil {
			return err
		}

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := utxo.UTXOSet{Blockchain: bc}
		err := UTXOSet.ReIndex()
		if err != nil {
			return err
		}
	}

	return nil
}

func handleInv(request []byte) error {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		err := sendGetData(payload.AddrFrom, "block", blockHash)
		if err != nil {
			return err
		}

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if !bytes.Equal(b, blockHash) {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			err := sendGetData(payload.AddrFrom, "tx", txID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handleGetBlocks(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload getBlocks

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	blocks, err := bc.GetBlockHashes()
	if err != nil {
		return err
	}

	return sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload getData

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return err
		}

		return sendBlock(payload.AddrFrom, block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		return SendTx(payload.AddrFrom, &tx)
	}

	return nil
}

func handleTx(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	txData := payload.Transaction
	tx, err := transactions.DeserializeTransaction(txData)
	if err != nil {
		return err
	}
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				err := sendInv(node, "tx", [][]byte{tx.ID})
				if err != nil {
					return err
				}
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*transactions.Transaction

			for id := range mempool {
				tx := mempool[id]
				valid, err := bc.VerifyTransaction(&tx)
				if err != nil {
					return err
				}
				if valid {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return nil
			}

			cbTx, err := transactions.NewCoinbaseTX(miningAddress, "")
			if err != nil {
				return err
			}

			txs = append(txs, cbTx)

			newBlock, err := bc.MineBlock(txs)
			if err != nil {
				return err
			}

			UTXOSet := utxo.UTXOSet{Blockchain: bc}
			UTXOSet.ReIndex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range KnownNodes {
				if node != nodeAddress {
					fmt.Println(node)
					err := sendInv(node, "block", [][]byte{newBlock.Hash})
					fmt.Println(err)
					if err != nil {
						return err
					}
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}

	return nil
}

func handleVersion(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload verzion

	fmt.Println("Got version request")

	buff.Write(request[commandLength:])
	err := msgpack.Unmarshal(buff.Bytes(), &payload)
	if err != nil {
		return err
	}

	myBestHeight, err := bc.GetBestHeight()
	fmt.Println(myBestHeight)
	if err != nil {
		return err
	}

	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		err := sendGetBlocks(payload.AddrFrom)
		if err != nil {
			return err
		}
	} else if myBestHeight > foreignerBestHeight {
		err := sendVersion(payload.AddrFrom, bc)
		if err != nil {
			return err
		}
	}

	if !nodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}

	for _, node := range KnownNodes {
		if node != KnownNodes[0] {
			err := sendAddr(node)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func StartServer(nodeID, minerAddress string) error {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	bc, err := blockchain.NewBlockchain()
	if err != nil {
		return err
	}

	if nodeAddress != KnownNodes[0] {
		err := sendVersion(KnownNodes[0], bc)
		if err != nil {
			return err
		}
	}

	fmt.Println("Server started")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn, bc)
	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
