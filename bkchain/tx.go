package bkchain

import (
	"bytes"
	"main/wallet"
	"encoding/gob"
  "log"
)


type Txinput struct {
  Ixid []byte
  Vout int
  Sig []byte
  Pubkey []byte
}

type Txoutputs struct{
  Outputs []Txoutput
}

type Txoutput struct {
  Value int
  Pubkeyhash []byte
}


func(in *Txinput) Useskey(pubkeyhash []byte)bool{
  lockinghash:=wallet.Pubkeyhash(in.Pubkey)
  return bytes.Compare(lockinghash, pubkeyhash)==0
}

func (out *Txoutput) Lock(address []byte){
  pubkeyhash:=wallet.Base58decode(address)
  pubkeyhash=pubkeyhash[1:len(pubkeyhash)-4]
  out.Pubkeyhash=pubkeyhash
}

func(out *Txoutput) Islockedwithkey(pubkeyhash []byte)bool{
  return bytes.Compare(out.Pubkeyhash, pubkeyhash)==0
}

func Newtxoutput(value int, address string) *Txoutput{
  txo:=&Txoutput{value, nil}
  txo.Lock([]byte(address))

  return txo
}

func (outs Txoutputs) Serialize() []byte{
  var buff bytes.Buffer

  enc:=gob.NewEncoder(&buff)
  err:=enc.Encode(outs)
  if err != nil{
    log.Panic(err)
  }
  return buff.Bytes()
}

func Deserializeoutputs(data []byte) Txoutputs{
  var outputs Txoutputs

  dec:=gob.NewDecoder(bytes.NewReader(data))
  err:=dec.Decode(&outputs)
  if err != nil {
    log.Panic(err)
  }
  return outputs
}
/*


// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
*/

