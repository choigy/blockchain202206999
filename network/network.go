package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
  "io"
	"log"
	"main/bkchain"
	"net"
)

const (
  protocol = "tcp"
  version = 1
  commandlength = 12
)

var (
  nodeaddress string
  mineraddress string
  Knownnodes = []string{"localhost:3000"}
  blocksintransit = [][]byte{}
  memorypool = make(map[string]bkchain.Transaction)
)

type Addr struct {
  Addrlist []string
}

type Block struct {
  Addrfrom string
  Block []byte
}

type Getblocks struct {
  Addrfrom string
}

type Getdata struct {
  Addrfrom string
  Type string
  Id []byte
}

type Inv struct {
  Addrfrom string
  Type string
  Items [][]byte
}

type Tx struct {
  Addrfrom string
  Transaction []byte
}

type Version struct {
  Version int
  Bestheight int
  Addrfrom string
}




func Cmdtobytes(cmd string) []byte {
  var bytes [commandlength]byte

  for i, c := range cmd {
    bytes[i] = byte(c)
  }
  return bytes[:]
}




func Bytestocmd(bytes []byte) string {
  var cmd []byte

  for _, b := range bytes {
    if b != 0x0 {
      cmd = append(cmd, b)
    }
  }
  return fmt.Sprintf("%s", cmd)
}



func Requestblocks(){
  for _, node := range Knownnodes {
    Sendgetblocks(node)
  }
}




func Handleconnection(conn net.Conn, bc *bkchain.Blockchain) {
  req, err := ioutil.ReadAll(conn)
  defer conn.Close()
  
  if err != nil {
    log.Panic(err)
  }
  cmd:=Bytestocmd(req[:commandlength])
  fmt.Printf("Received %s command\n", cmd)

  switch cmd {
	case "addr":
		Handleaddr(req)
	case "block":
		Handleblock(req, bc)
	case "inv":
		Handleinv(req, bc)
	case "getblocks":
		Handlegetblocks(req, bc)
	case "getdata":
		Handlegetdata(req, bc)
	case "tx":
		Handletx(req, bc)
	case "version":
		Handleversion(req, bc)
	default:
		fmt.Println("Unknown command!")
	}
}


func Extractcmd(req []byte) []byte {
  return req[:commandlength]
}



func Sendaddr(address string) {
  nodes:=Addr{Knownnodes}
  nodes.Addrlist = append(nodes.Addrlist, nodeaddress)
  payload:=Gobencode(nodes)
  request:=append(Cmdtobytes("addr"), payload...)

  Senddata(address, request)
}






func Sendblock(addr string, b *bkchain.Block) {
  data:=Block{nodeaddress, b.Serialize()}
  payload:=Gobencode(data)
  request:=append(Cmdtobytes("block"), payload...)

  Senddata(addr, request)
}






func Sendinv(address, kind string, items [][]byte){
  inventory:=Inv{nodeaddress, kind, items}
  payload:=Gobencode(inventory)
  request:=append(Cmdtobytes("inv"), payload...)

  Senddata(address, request)
}






func Sendtx(addr string, tx *bkchain.Transaction){
  data:=Tx{nodeaddress, tx.Serialize()}
  payload:=Gobencode(data)
  request:=append(Cmdtobytes("tx"), payload...)

  Senddata(addr, request)
}







func Sendversion(addr string, bc *bkchain.Blockchain) {
  bestheight:=bc.Getbestheight()
  payload:=Gobencode(Version{version, bestheight, nodeaddress})
  request:=append(Cmdtobytes("version"), payload...)

  Senddata(addr, request)
}







func Sendgetblocks(address string) {
  payload:=Gobencode(Getblocks{nodeaddress})
  request:=append(Cmdtobytes("getblocks"), payload...)

  Senddata(address, request)
}






func Sendgetdata(address, kind string, id []byte) {
  payload:=Gobencode(Getdata{nodeaddress, kind, id})
  request:=append(Cmdtobytes("getdata"), payload...)

  Senddata(address, request)
}







func Handleaddr(request []byte){
  var buff bytes.Buffer
  var payload Addr

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }

  Knownnodes = append(Knownnodes, payload.Addrlist...)
  fmt.Printf("There are %d known nodes now!\n", len(Knownnodes))
  Requestblocks()
}







func Handleblock(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Block

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }

  blockData:=payload.Block
  block:=bkchain.DeSerializeBlock(blockData)

  fmt.Println("Recevied a new block")
  bc.AddBlock(block)

  fmt.Printf("Added block %x\n", block.Hash)

  if len(blocksintransit)>0{
    blockhash:=blocksintransit[0]
    Sendgetdata(payload.Addrfrom, "block", blockhash)

    blocksintransit = blocksintransit[1:]
  } else {
    UTXOSet := bkchain.UTXOSet{bc}
    UTXOSet.Reindex()
  }
}




func Handlegetblocks(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Getblocks

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }

  blocks := bc.Getblockhashes()
  Sendinv(payload.Addrfrom, "block", blocks)
}




func Handlegetdata(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Getdata

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }
  if payload.Type == "block" {
    block, err := bc.GetBlock([]byte(payload.Id))
    if err != nil {
      return
    }
    Sendblock(payload.Addrfrom, &block)
  }

  if payload.Type == "tx" {
    txid := hex.EncodeToString(payload.Id)
    tx := memorypool[txid]

    Sendtx(payload.Addrfrom, &tx)
  }
}




func Handleversion(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Version

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(payload)
  if err != nil {
    log.Panic(err)
  }

  bestheight := bc.Getbestheight()
  otherheight:=payload.Bestheight

  if bestheight < otherheight {
    Sendgetblocks(payload.Addrfrom)
  } else if bestheight > otherheight {
    Sendversion(payload.Addrfrom, bc)
  }

  if !Nodeisknown(payload.Addrfrom) {
    Knownnodes = append(Knownnodes, payload.Addrfrom)
  }
}




func Handletx(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Tx

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }

  txdata:=payload.Transaction
  tx:=bkchain.Deserializetx(txdata)
  memorypool[hex.EncodeToString(tx.Id)] = tx

  if nodeaddress == Knownnodes[0] {
    for _, node := range Knownnodes {
      if node != nodeaddress && node != payload.Addrfrom {
        Sendinv(node, "tx", [][]byte{tx.Id})
      }
    }
  } else {
    if len(memorypool) >= 2 && len(mineraddress)>0 {
      Minetx(bc)
    }
  }
}

func Minetx(bc *bkchain.Blockchain){
  var txs []*bkchain.Transaction

  for id := range memorypool {
    tx := memorypool[id]
    if bc.Verifytransaction(&tx) {
      txs = append(txs, &tx)
    }
  }
  if len(txs)==0{
    fmt.Println("All transactions are invalid!")
    return
  }
  cbtx:=bkchain.Newcoinbasetx(mineraddress, "")
  txs = append(txs, cbtx)

  newblock:= bc.Mineblock(txs)
  UTXOSet:=bkchain.UTXOSet{bc}
  UTXOSet.Reindex()

  fmt.Println("New block is mined!")

  for _, tx := range txs {
    txid := hex.EncodeToString(tx.Id)
    delete(memorypool, txid)
  }

  for _, node := range Knownnodes {
    if node != nodeaddress {
      Sendinv(node, "block", [][]byte{newblock.Hash})
    }
  }
  if len(memorypool)>0 {
    Minetx(bc)
  }
}


func Handleinv(request []byte, bc *bkchain.Blockchain) {
  var buff bytes.Buffer
  var payload Inv

  buff.Write(request[commandlength:])
  dec:=gob.NewDecoder(&buff)
  err:=dec.Decode(&payload)
  if err != nil {
    log.Panic(err)
  }

  fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

  if payload.Type=="block"{
    blocksintransit=payload.Items

    blockhash:=payload.Items[0]
    Sendgetdata(payload.Addrfrom, "block", blockhash)

    newintransit := [][]byte{}
    for _, b := range blocksintransit {
      if bytes.Compare(b, blockhash) != 0 {
        newintransit = append(newintransit, b)
      }
    }
    blocksintransit=newintransit
  }
  if payload.Type == "tx" {
    txid := payload.Items[0]

    if memorypool[hex.EncodeToString(txid)].Id == nil {
      Sendgetdata(payload.Addrfrom, "tx", txid)
    }
  }
}




func Startserver(nodeid, mineraddress string) {
  nodeaddress = fmt.Sprintf("localhost:%s", nodeid)
  mineraddress = mineraddress
  ln, err := net.Listen(protocol, nodeaddress)
  if err != nil {
    log.Panic(err)
  }
  defer ln.Close()

  bc:=bkchain.ContinueBlockchain(nodeid)
  defer bc.Db.Close()
  //go Closedb(bc)
  
  if nodeaddress != Knownnodes[0] {
    Sendversion(Knownnodes[0], bc)
  }
  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Panic(err)
    }
    go Handleconnection(conn, bc)
  }
}
/*	
*/




func Senddata(addr string, data []byte){
  conn, err := net.Dial(protocol,addr)
  if err != nil {
    fmt.Printf("%s is not available\n", addr)
    var updatenodes []string

    for _, node := range Knownnodes {
      if node != addr {
        updatenodes = append(updatenodes, node)
      }
    }
    Knownnodes=updatenodes
    return 
  }
  defer conn.Close()

  _, err = io.Copy(conn, bytes.NewReader(data))
  if err != nil {
    log.Panic(err)
  }
}





func Gobencode(data interface{}) []byte {
  var buff bytes.Buffer

  enc:=gob.NewEncoder(&buff)
  err:=enc.Encode(data)
  if err != nil {
    log.Panic(err)
  }
  return buff.Bytes()
}





func Nodeisknown(addr string) bool {
  for _, node := range Knownnodes{
    if node == addr {
      return true
    }
  }
  return false
}
