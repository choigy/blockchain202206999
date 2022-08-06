package main

import (
	//"fmt"
  //"flag"
  "os"
  //"log"
	//"strconv"
  //"bytes"
  //"time"
  //"encoding/binary"
  //"crypto/sha256"
  //"math/big"
  //"math"
  //"log"
  //"encoding/hex"
  //"main/bkchain"
  "main/cli"
  //"main/wallet"
)


func main() {
  defer os.Exit(1)
  cmd:=cli.CLI{}
  cmd.Run()
  //w:=wallet.Makewallet()
  //w.Getaddress()
}



