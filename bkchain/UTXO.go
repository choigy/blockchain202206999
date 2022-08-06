package bkchain

import (
  "encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)



type UTXOSet struct {
  Blockchain *Blockchain
}



func (u UTXOSet) Reindex(){
  db:=u.Blockchain.Db
  bucketname:= []byte("utxobucket")

  err:=db.Update(func(tx *bolt.Tx)error{
    err:=tx.DeleteBucket(bucketname)
    if err != nil && err != bolt.ErrBucketNotFound{
      log.Panic(err)
    }

    _, err=tx.CreateBucket(bucketname)
    if err != nil {
      log.Panic(err)
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }

  UTXO := u.Blockchain.FindUTXO()
  err = db.Update(func(tx *bolt.Tx)error{
    b:=tx.Bucket(bucketname)

    for txid, outs := range UTXO {
      key, err := hex.DecodeString(txid)
      if err != nil {
        log.Panic(err)
      }

      err=b.Put(key, outs.Serialize())
      if err != nil {
        log.Panic(err)
      }
    }
    return nil
  })
}
  





func (u UTXOSet) Update(block *Block) {
  bucketname:=[]byte("utxobucket")
  db:=u.Blockchain.Db

  err:= db.Update(func(tx *bolt.Tx)error{
    b:=tx.Bucket(bucketname)

    for _, tx := range block.Transactions{
      if tx.IsCoinbase() == false{
        for _, in := range tx.Vin {
          updateoutputs:=Txoutputs{}
          outbytes:=b.Get(in.Ixid)
          outs:=Deserializeoutputs(outbytes)

          for idx, out := range outs.Outputs{
            if idx != in.Vout{
              updateoutputs.Outputs = append(updateoutputs.Outputs, out)
            }
          }
          if len(updateoutputs.Outputs)==0{
            err:=b.Delete(in.Ixid)
            if err != nil {
              log.Panic(err)
            }
          } else {
            err := b.Put(in.Ixid, updateoutputs.Serialize())
            if err != nil {
              log.Panic(err)
            }
          }
          
        }
      }
      newoutputs:=Txoutputs{}
      for _, out := range tx.Vout{
        newoutputs.Outputs = append(newoutputs.Outputs, out)
      }
      err:=b.Put(tx.Id, newoutputs.Serialize())
      if err != nil {
        log.Panic(err)
      }
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
}






func (u UTXOSet) Counttxs() int {
  bucketname:=[]byte("utxobucket")
  db:=u.Blockchain.Db
  counter:=0

  err:= db.Update(func(tx *bolt.Tx)error {
    b:=tx.Bucket(bucketname)
    c:=b.Cursor()

    for key, _ := c.First(); key!=nil; key, _ = c.Next(){
      counter++
    }

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return counter
}





func (u UTXOSet) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int){
  bucketname:=[]byte("utxobucket")
  unspentoutputs:=make(map[string][]int)
  acc:=0
  db:=u.Blockchain.Db

  err:= db.View(func(tx *bolt.Tx)error{
    b:=tx.Bucket(bucketname)
    c:=b.Cursor()

    for k, v:=c.First(); k!=nil; k,v=c.Next(){
      txid:=hex.EncodeToString(k)
      outs:=Deserializeoutputs(v)
      
      for idx, out := range outs.Outputs{
        if out.Islockedwithkey(pubkeyhash) && acc<amount {
          acc+=out.Value
          unspentoutputs[txid]=append(unspentoutputs[txid], idx)
        }
      }
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return acc, unspentoutputs
}






func (u UTXOSet) FindUTXO(pubkeyhash []byte) []Txoutput {
  bucketname:=[]byte("utxobucket")
  var UTSOs []Txoutput
  db:=u.Blockchain.Db

  err:=db.View(func(tx *bolt.Tx)error{
    b:=tx.Bucket(bucketname)
    c:=b.Cursor()

    for k, v := c.First(); k!=nil; k, v = c.Next(){
      outs:=Deserializeoutputs(v)

      for _, out := range outs.Outputs{
        if out.Islockedwithkey(pubkeyhash){
          UTSOs=append(UTSOs, out)
        }
      }
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return UTSOs
}

/*
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

*/

