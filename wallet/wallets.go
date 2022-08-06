package wallet

import(
  "bytes"
  "crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct{
  Wallets map[string]*Wallet
}

func CreateWallets(nodeid string)(*Wallets, error){
  wallets:=Wallets{}
  wallets.Wallets=make(map[string]*Wallet)
  err:=wallets.LoadFile(nodeid)
  return &wallets, err
}

func (ws Wallets)Getwallet(address string)Wallet{
  return *ws.Wallets[address]
}

func (ws *Wallets)Getalladdress()[]string{
  var alladdress []string
  for address:=range ws.Wallets{
    alladdress=append(alladdress, address)
  }
  return alladdress
}

func (ws *Wallets)Addwallet()string{
  wallet:=Makewallet()
  address:=fmt.Sprintf("%s",wallet.Getaddress())
  ws.Wallets[address]=wallet
  return address
}

func (ws *Wallets)LoadFile(nodeid string)error{
  walletFile := fmt.Sprintf(walletFile, nodeid)
  
  if _,err:=os.Stat(walletFile); os.IsNotExist(err){
    return err
  }
  
  var wallets Wallets
  filecontent, err := ioutil.ReadFile(walletFile)
  if err != nil {
    log.Panic(err)
  }
  gob.Register(elliptic.P256())
  decoder:=gob.NewDecoder(bytes.NewReader(filecontent))
  err=decoder.Decode(&wallets)
  if err != nil {
    log.Panic(err)
  }
  ws.Wallets=wallets.Wallets
  return nil
}

func (ws *Wallets)Savefile(nodeid string){
  var content bytes.Buffer
  walletFile := fmt.Sprintf(walletFile, nodeid)
  
  gob.Register(elliptic.P256())
  encoder:=gob.NewEncoder(&content)
  err:=encoder.Encode(ws)
  if err != nil {
    log.Panic(err)
  }
  err=ioutil.WriteFile(walletFile, content.Bytes(), 0644)
  if err != nil {
    log.Panic(err)
  }
}