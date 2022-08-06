package wallet

import (
	"crypto/ecdsa"
  "crypto/elliptic"
  "crypto/rand"
	"crypto/sha256"
	"log"
  "fmt"
  "bytes"

	"golang.org/x/crypto/ripemd160"
)


const version = byte(0x00)
const checksumlenth=4

type Wallet struct {
  PrivateKey ecdsa.PrivateKey
  PublicKey []byte
}

func Makewallet () *Wallet{
  prikey, pubkey := Newkeypair()
  wallet:=Wallet{prikey, pubkey}
  return &wallet
}

func (w Wallet)Getaddress()[]byte{
  pubkeyhash:=Pubkeyhash(w.PublicKey)
  versionpayload:=append([]byte{version}, pubkeyhash...)
  checksum:=Checksum(versionpayload)
  fullpayload:=append(versionpayload, checksum...)
  address:=Base58encode(fullpayload)

  fmt.Printf("pubkey:%x\n",w.PublicKey)
  fmt.Printf("pubkeyhash:%x\n",pubkeyhash)
  fmt.Printf("address:%x\n",address)
  return address
}


func Validateaddress(address string) bool {
  pubkeyhash:=Base58decode([]byte(address))
  actualchecksum:=pubkeyhash[len(pubkeyhash)-checksumlenth:]
  version:=pubkeyhash[0]
  pubkeyhash=pubkeyhash[1:len(pubkeyhash)-checksumlenth]
  targetchecksum:=Checksum(append([]byte{version},pubkeyhash...))

  return bytes.Compare(actualchecksum, targetchecksum)==0
}





func Newkeypair()(ecdsa.PrivateKey, []byte){
  curve:=elliptic.P256()
  prikey, err := ecdsa.GenerateKey(curve, rand.Reader)
  if err != nil{
    log.Panic(err)
  }
  pubkey:=append(prikey.PublicKey.X.Bytes(), prikey.PublicKey.Y.Bytes()...)
  return *prikey, pubkey
}

func Pubkeyhash(pubkey []byte)[]byte{
  pubhash:=sha256.Sum256(pubkey)
  hasher:=ripemd160.New()
  _, err:=hasher.Write(pubhash[:])
  if err != nil {
    log.Panic(err)
  }
  pubripemd:=hasher.Sum(nil)
  return pubripemd
}

func Checksum(payload []byte)[]byte{
  first:=sha256.Sum256(payload)
  second:=sha256.Sum256(first[:])
  return second[:checksumlenth]
}

