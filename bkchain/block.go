package bkchain

import (
	"bytes"
	//"crypto/sha256"

	//"strconv"
	"encoding/gob"
	//"encoding/hex"
	"log"
	"time"

	//"github.com/boltdb/bolt"
	//"fmt"
)

type Block struct {
  Timestamp     int64
  Transactions  []*Transaction
  PrevBlockHash []byte
  Hash          []byte
  nonce         int
  Height        int
}



/*
func (b *Block)SetHash(){
  timestamp:=[]byte(strconv.FormatInt(b.Timestamp, 10))
  headers:=bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
  hash:=sha256.Sum256(headers)
  b.Hash=hash[:]
}
*/

func NewBlock(transactions []*Transaction, prevBlockhash []byte, height int) *Block {
  block:=&Block{time.Now().Unix(), transactions, prevBlockhash, []byte{}, 0, height}
  pow:=NewProofOfWork(block)
  nonce, hash := pow.Run()
  block.Hash=hash[:]
  block.nonce=nonce
  return block
}



func GenesisBlock(coinbase *Transaction) *Block {
  return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
  
}



func (b *Block)HashTransactions()[]byte{
  var txHashes [][]byte
  

  for _, tx:=range b.Transactions{
    txHashes=append(txHashes, tx.Serialize())
  }

  tree:=Newmerkletree(txHashes)
  return tree.Rootnode.Data
}





func (b *Block) Serialize() []byte {
  var result bytes.Buffer
  encoder:=gob.NewEncoder(&result)
  err:=encoder.Encode(b)
  if err != nil {
		log.Fatal(err)
	}
  return result.Bytes()
}

func DeSerializeBlock(d []byte)*Block {
  var block Block
  decoder:=gob.NewDecoder(bytes.NewReader(d))
  err:=decoder.Decode(&block)
  if err != nil {
		log.Fatal(err)
	}
  return &block
}



