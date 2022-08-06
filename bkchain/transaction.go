package bkchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"main/wallet"
	"math/big"
  "crypto/rand"
  "strings"
)

type Transaction struct {
  Id []byte
  Vin []Txinput
  Vout []Txoutput
}



func NewTransaction(fwallet *wallet.Wallet, to string, amount int, UTXO *UTXOSet)*Transaction{
  var inputs []Txinput
  var outputs []Txoutput

  
  pubkeyhash:=wallet.Pubkeyhash(fwallet.PublicKey)
  
  acc, validOutputs := UTXO.FindSpendableOutputs(pubkeyhash, amount)

  if acc<amount{
    log.Panic("Fund is not enough")
  }

  for txid, outs := range validOutputs{
    txId, err:=hex.DecodeString(txid)
    if err != nil{
      log.Panic("decoding error")
    }
    for _, out:=range outs{
      input:=Txinput{txId, out, nil, fwallet.PublicKey}
      inputs=append(inputs, input)
    }
  }

  from:=fmt.Sprintf("%s",fwallet.Getaddress())
  outputs=append(outputs, *Newtxoutput(amount, to))
  if acc>amount{
    outputs=append(outputs, *Newtxoutput(acc-amount, from))
  }
  tx:=Transaction{nil, inputs, outputs}
  tx.Id=tx.Hash()
  UTXO.Blockchain.Signtransaction(&tx, fwallet.PrivateKey)
  //Signtansaction(&tx, w.PrivateKey)Signtransaction
  return &tx
}




func (tx *Transaction) Hash() []byte {
  var hash [32]byte

  txcopy:=*tx
  txcopy.Id=[]byte{}

  hash=sha256.Sum256(txcopy.Serialize())
  return hash[:]
}




func (tx Transaction) Serialize() []byte{
  var encoded bytes.Buffer

  enc:=gob.NewEncoder(&encoded)
  err:=enc.Encode(tx)
  if err != nil {
    log.Panic(err)
  }

  return encoded.Bytes()
}




func (tx *Transaction)SetId(){
  var encoded bytes.Buffer
  var hash [32]byte

  encoder:=gob.NewEncoder(&encoded)
  err:=encoder.Encode(tx)

  if err != nil {
    log.Fatal(err)
  }
  hash=sha256.Sum256(encoded.Bytes())
  tx.Id=hash[:]
}





func Newcoinbasetx(to, data string) *Transaction{
  if data==""{
    randdata:=make([]byte, 24)
    _, err:=rand.Read(randdata)
    if err != nil {
      log.Panic(err)
    }
    data = fmt.Sprintf("%x", randdata)
  }
  txin:=Txinput{[]byte{}, -1, nil, []byte(data)}
  txout:=Newtxoutput(20, to)
  tx:=Transaction{nil, []Txinput{txin}, []Txoutput{*txout}}
  tx.SetId()
  return &tx
}





func (tx *Transaction)IsCoinbase()bool{
  return len(tx.Vin)==1&&len(tx.Vin[0].Ixid)==0&&tx.Vin[0].Vout==-1
}





func(tx *Transaction)Sign(privkey ecdsa.PrivateKey, prevtxs map[string]Transaction){
  if tx.IsCoinbase(){
    return
  }
  for _, input := range tx.Vin {
    if prevtxs[hex.EncodeToString(input.Ixid)].Id==nil{
      log.Panic("previous tx does not exist")
    }
  }
  txcopy:=tx.Trimmedcopy()
  for idx, in := range txcopy.Vin{
    prevtx:=prevtxs[hex.EncodeToString(in.Ixid)]
    txcopy.Vin[idx].Sig=nil
    txcopy.Vin[idx].Pubkey=prevtx.Vout[in.Vout].Pubkeyhash
    txcopy.Id=txcopy.Hash()
    txcopy.Vin[idx].Pubkey=nil

    r, s, err := ecdsa.Sign(rand.Reader, &privkey, txcopy.Id)
    if err != nil {
      log.Panic(err)
    }
    sig:=append(r.Bytes(), s.Bytes()...)
    txcopy.Vin[idx].Sig=sig
  } 
}





func(tx *Transaction) Trimmedcopy() Transaction {
  var inputs []Txinput
  var outputs []Txoutput

  for _, in := range tx.Vin{
    inputs = append(inputs, Txinput{in.Ixid, in.Vout, nil, nil})
  }

  for _, out := range tx.Vout{
    outputs = append(outputs, Txoutput{out.Value, out.Pubkeyhash})
  }

  txcopy:=Transaction{tx.Id, inputs, outputs}
  return txcopy
}






func(tx *Transaction)Verify(prevtxs map[string]Transaction)bool{
  if tx.IsCoinbase(){
    return true
  }
  for _, in := range tx.Vin{
    if prevtxs[hex.EncodeToString(in.Ixid)].Id==nil {
      log.Panic("previous tx does not exist")
    }
  }

  txcopy:=tx.Trimmedcopy()
  curve:=elliptic.P256()

  for idx, in := range txcopy.Vin{
    prevtx:=prevtxs[hex.EncodeToString(in.Ixid)]
    txcopy.Vin[idx].Sig=nil
    txcopy.Vin[idx].Pubkey=prevtx.Vout[in.Vout].Pubkeyhash
    txcopy.Id=txcopy.Hash()
    txcopy.Vin[idx].Pubkey=nil

    r:=big.Int{}
    s:=big.Int{}
    siglen:=len(in.Sig)
    r.SetBytes(in.Sig[:(siglen/2)])
    s.SetBytes(in.Sig[(siglen/2):])

    x:=big.Int{}
    y:=big.Int{}
    keylen:=len(in.Pubkey)
    x.SetBytes(in.Pubkey[:(keylen/2)])
    y.SetBytes(in.Pubkey[(keylen/2):])

    rawpubkey:=ecdsa.PublicKey{curve, &x, &y}
    if ecdsa.Verify(&rawpubkey, txcopy.Id, &r, &s)==false{
      return false
    }
  }
  return true
}






func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.Id))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       Txinput->Ixid:      %x", input.Ixid))
		lines = append(lines, fmt.Sprintf("       Txinput->Vout:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Txinput->Signature: %x", input.Sig))
		lines = append(lines, fmt.Sprintf("       Txinput->PubKey:    %x", input.Pubkey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Txoutput->Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Txoutput->Pubkeyhash: %x", output.Pubkeyhash))
	}

	return strings.Join(lines, "\n")
}




func Deserializetx(data []byte) Transaction {
  var transaction Transaction

  dec:=gob.NewDecoder(bytes.NewReader(data))
  err:=dec.Decode(&transaction)
  if err != nil {
    log.Panic(err)
  }
  return transaction
}

/*

	func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

*/



