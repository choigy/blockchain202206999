package bkchain

import (
	//"bytes"
	//"crypto/sha256"

	//"strconv"
	//"encoding/gob"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"log"
	"os"

	//"time"

	"github.com/boltdb/bolt"
	"fmt"
)

const dbfile = "blockchain.db"
const blocksbucket = "blocks"
const genesiscbdata = "First transaction"

type Blockchain struct{
  tip []byte
  Db *bolt.DB
}

type BlockchainIterator struct{
  currenthash []byte
  Db *bolt.DB
}


func(bc *Blockchain) Iterator() *BlockchainIterator{
  bci:=&BlockchainIterator{bc.tip, bc.Db}
  return bci 
}


/*
func (b *Block)SetHash(){
  timestamp:=[]byte(strconv.FormatInt(b.Timestamp, 10))
  headers:=bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
  hash:=sha256.Sum256(headers)
  b.Hash=hash[:]
}
*/



func (bc *Blockchain) AddBlock(block *Block) {
  err:=bc.Db.Update(func(tx *bolt.Tx)error {
    b:=tx.Bucket([]byte(blocksbucket))
    blockindb := b.Get(block.Hash)

    if blockindb != nil {
      return nil 
    }

    blockdata := block.Serialize()
    err:=b.Put(block.Hash, blockdata)
    if err != nil {
      log.Panic(err)
    }
    
    lasthash:=b.Get([]byte("1"))
    lastblockdata := b.Get(lasthash)
    lastblock:=DeSerializeBlock(lastblockdata)

    if block.Height > lastblock.Height {
      err:=b.Put([]byte("1"), block.Hash)
      if err != nil {
        log.Panic(err)
      }
      bc.tip = block.Hash
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
}






func Createblockchain(address, nodeid string)*Blockchain {
  dbfile := fmt.Sprintf(dbfile, nodeid)
	if Dbexist(dbfile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

  
  var tip []byte
  Db, err := bolt.Open(dbfile,0600,nil)
  err=Db.Update(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte(blocksbucket))
    if b==nil{
      cobatx:=Newcoinbasetx(address, "")      
      genesis:=GenesisBlock(cobatx)
      b, err:=tx.CreateBucket([]byte("blocksbucket"))
      err=b.Put(genesis.Hash, genesis.Serialize())
      err=b.Put([]byte("1"), genesis.Hash)
      if err != nil {
		    log.Fatal(err)
	    }
      tip=genesis.Hash
    }else{
      tip=b.Get([]byte("1"))
    }
    return nil
  })
  if err != nil {
		log.Fatal(err)
	}
  bc:=Blockchain{tip, Db}
  return &bc
}





func ContinueBlockchain(nodeid string)*Blockchain{
  dbfile:=fmt.Sprintf(dbfile,nodeid)
  if Dbexist(dbfile)==false {
    fmt.Println("No existing blockchain found.")
    os.Exit(1)
  }
  var tip []byte
  db, err:=bolt.Open(dbfile, 0600, nil)
  err=db.Update(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte(blocksbucket))
    tip=b.Get([]byte("1"))
    return nil
  })
  if err != nil{
    log.Panic(err)
  }
  bc:=Blockchain{tip, db}
  return &bc
}






func (bc *Blockchain) GetBlock(blockhash []byte) (Block, error) {
  var block Block

  err:=bc.Db.View(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte(blocksbucket))

    blockdata:=b.Get(blockhash)
    if blockdata == nil{
      return errors.New("Block is not exist")
    }

    block=*DeSerializeBlock(blockdata)

    return nil
  })
  if err != nil {
    return block, err
  }
  return block, nil
}







func (bc *Blockchain) Getblockhashes() [][]byte {
  var blocks [][]byte
  bci:=bc.Iterator()

  for {
    block:=bci.Next()
    blocks = append(blocks, block.Hash)
    if len(block.PrevBlockHash)==0{
      break
    }
  }
  return blocks
}






func (bc *Blockchain) Getbestheight() int {
  var lastblock Block

  err:= bc.Db.View(func(tx *bolt.Tx)error {
    b:=tx.Bucket([]byte(blocksbucket))
    lasthash:=b.Get([]byte("1"))
    blockdata:=b.Get(lasthash)
    lastblock = *DeSerializeBlock(blockdata)

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return lastblock.Height
}





func (bc *Blockchain)Mineblock (transactions []*Transaction) *Block {
  var lasthash []byte
  var lastheight int

  for _, tx := range transactions {
    if bc.Verifytransaction(tx) != true {
      log.Panic("error: invalid transaction")
    }
  }

  err := bc.Db.View(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte(blocksbucket))
    lasthash = b.Get([]byte("1"))

    blockdata:=b.Get(lasthash)
    block:=DeSerializeBlock(blockdata)

    lastheight = block.Height

    return nil
  })
  if err != nil {
    log.Panic(err)
  }

  newblock:=NewBlock(transactions, lasthash, lastheight+1)

  err=bc.Db.Update(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte(blocksbucket))
    err:=b.Put(newblock.Hash, newblock.Serialize())
    if err != nil {
      log.Panic(err)
    }

    err=b.Put([]byte("1"), newblock.Hash)
    if err != nil {
      log.Panic(err)
    }

    bc.tip=newblock.Hash

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return newblock
}
/*
*/




func (bc *Blockchain)FindUnspentTransactions(pubkeyhash []byte) []Transaction{
  var unspentTXs []Transaction
  spentTXOs:=make(map[string][]int)
  bci:=bc.Iterator()

  for{
    block:=bci.Next()
    for _, tx := range block.Transactions {
      txid:= hex.EncodeToString(tx.Id)
    outputs:
      for outIdx, out := range tx.Vout{
        if spentTXOs[txid] != nil {
          for _, spentout := range spentTXOs[txid]{
            if spentout==outIdx{
              continue outputs
            }
          }
        }
        if out.Islockedwithkey(pubkeyhash){
          unspentTXs=append(unspentTXs, *tx)
        }
      }
      if tx.IsCoinbase()==false{
        for _, in := range tx.Vin {
          if in.Useskey(pubkeyhash){
            intxid:=hex.EncodeToString(in.Ixid)
            spentTXOs[intxid]=append(spentTXOs[intxid], in.Vout)
          }
        }
      }
    }
    if len(block.PrevBlockHash)==0{
      break
    }
  }
  return unspentTXs
}

/*
func (bc *Blockchain)FindUTXO(pubkeyhash []byte) []Txoutput{
  var UTXOs []Txoutput
  unspentTransactions:=bc.FindUnspentTransactions(pubkeyhash)
  for _, tx := range unspentTransactions{
    for _, out:=range tx.Vout {
      if out.Islockedwithkey(pubkeyhash){
        UTXOs=append(UTXOs, out)
      }
    }
  }
  return UTXOs
}
*/

func (bc *Blockchain) FindUTXO() map[string]Txoutputs{
  UTXO := make(map[string]Txoutputs)
  spentTXOs:=make(map[string][]int)
  bci:=bc.Iterator()

  for {
    block:=bci.Next()

    for _, tx := range block.Transactions{
      txid:= hex.EncodeToString(tx.Id)

    Outputs:
      for idx, out := range tx.Vout {
        if spentTXOs[txid] != nil {
          for _, spentoutidx := range spentTXOs[txid]{
            if spentoutidx == idx {
              continue Outputs
            }
          }
        }
        outs:=UTXO[txid]
        outs.Outputs = append(outs.Outputs, out)
        UTXO[txid]=outs
      }

      if tx.IsCoinbase()==false {
        for _, in := range tx.Vin{
          intxid:=hex.EncodeToString(in.Ixid)
          spentTXOs[intxid] = append(spentTXOs[intxid], in.Vout)
        }
      }
    }
    if len(block.PrevBlockHash)==0{
      break
    }
  }
  return UTXO
}



func (bc *Blockchain)FindSpendableOutputs(pubkeyhash []byte, amount int)(int, map[string][]int){
  unspentoutputs:=make(map[string][]int)
  unspentTxs:=bc.FindUnspentTransactions(pubkeyhash)
  accumulated:=0

  work:
  for _, tx := range unspentTxs{
    txId:=hex.EncodeToString(tx.Id)
    for outIdx, out := range tx.Vout{
      if out.Islockedwithkey(pubkeyhash)&&accumulated<amount{
        accumulated+=out.Value
        unspentoutputs[txId]=append(unspentoutputs[txId], outIdx)
      }
      if accumulated>=amount{
        break work
      }
    }
  }
  return accumulated, unspentoutputs 
}





func (i *BlockchainIterator)Next() *Block{
  var block *Block
  err:=i.Db.View(func(tx *bolt.Tx)error{
    b:=tx.Bucket([]byte("blocksbucket"))
    encodedhash:=b.Get(i.currenthash)
    block=DeSerializeBlock(encodedhash)
    return nil
  })
  if err != nil {
		log.Fatal(err)
	}
  i.currenthash=block.PrevBlockHash
  return block
}



func (bc *Blockchain)Findtransaction(txid []byte) (Transaction, error){
  bci:=bc.Iterator()
  for {
    block:=bci.Next()

    for _, tx := range block.Transactions{
      if bytes.Compare(txid, tx.Id)==0{
        return *tx, nil
      }
    }
    if len(block.PrevBlockHash)==0{
      break
    }
  }
  return Transaction{}, errors.New("tx does not exist")
}



func (bc *Blockchain)Signtransaction(tx *Transaction, privkey ecdsa.PrivateKey){
  prevtxs:=make(map[string]Transaction)

  for _, in := range tx.Vin{
    prevtx, err:=bc.Findtransaction(in.Ixid)
    if err != nil {
      log.Panic(err)
    }
    prevtxs[hex.EncodeToString(prevtx.Id)]=prevtx
  }
  tx.Sign(privkey, prevtxs)
}




func (bc *Blockchain) Verifytransaction(tx *Transaction)bool{
  if tx.IsCoinbase(){
    return true
  }
  prevtxs:=make(map[string]Transaction)

  for _, in := range tx.Vin{
    prevtx, err := bc.Findtransaction(in.Ixid)
    if err != nil {
      log.Panic(err)
    }
    prevtxs[hex.EncodeToString(prevtx.Id)]=prevtx
  }
  return tx.Verify(prevtxs)
}





func Dbexist(dbfile string) bool {
  if _, err := os.Stat(dbfile); os. IsNotExist(err) {
    return false
  }

  return true
}
