package cli

import (
	"fmt"
  "flag"
  "os"
  "log"
	"strconv"
  //"bytes"
  //"time"
  //"encoding/binary"
  //"crypto/sha256"
  //"math/big"
  //"math"
  //"log"
  //"encoding/hex"
  "main/bkchain"
  "main/wallet"
  "main/network"
)

type CLI struct {}


func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  newblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
  fmt.Println("  createwallet -Create a new wallet")
  fmt.Println("  listaddress -List all address")
  fmt.Println("  reindexutxo -Rebuilds the UTXO set")
  fmt.Println("  startnode -miner Address -Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}


func (cli *CLI)validateArgs(){
  if len(os.Args)<2{
    cli.printUsage()
    os.Exit(1)
  }
}

func (cli *CLI) Run () {
  cli.validateArgs()

  nodeid := os.Getenv("NODE_ID")
  if nodeid == ""{
    fmt.Printf("NODE_ID env. var is not set")
    os.Exit(1)
  }
  
  printchainCmd:=flag.NewFlagSet("printchain", flag.ExitOnError)
  getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
  sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
  createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
  listAddressesCmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
  reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
  startnodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	

  getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
  sendMine := sendCmd.Bool("mine", false, "mine immediately")
  startNodeMiner := startnodeCmd.String("miner", "", "Enable mining mode and send reward")
	

  switch os.Args[1] {
  case "reindexutxo":
    err:=reindexUTXOCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "getbalance":
    err:=getBalanceCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "createblockchain":
    err:=createBlockchainCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "send":
    err:=sendCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "startnode":
    err:=startnodeCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}    
  case "printchain":
    err:=printchainCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "createwallet":
    err:=createWalletCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}
  case "listaddress":
    err:=listAddressesCmd.Parse(os.Args[2:])
    if err != nil {
		log.Panic(err)
	}  
  default:
    os.Exit(1)
  }

  if getBalanceCmd.Parsed() {
    if *getBalanceAddress==""{
      getBalanceCmd.Usage()
      os.Exit(1)
    }
    cli.getBalance(*getBalanceAddress, nodeid)
  }

  if printchainCmd.Parsed(){
    cli.printchain(nodeid)
  }

  if createWalletCmd.Parsed(){
    cli.createwallet(nodeid)
  }

  if listAddressesCmd.Parsed(){
    cli.listaddresses(nodeid)
  }

  if reindexUTXOCmd.Parsed(){
    cli.reindexUTXO(nodeid)
  }
  
  if createBlockchainCmd.Parsed(){
    if *createBlockchainAddress==""{
      createBlockchainCmd.Usage()
      os.Exit(1)
    }
    cli.createBlockchain(*createBlockchainAddress, nodeid)
  }

  if sendCmd.Parsed(){
    if *sendFrom==""||*sendTo==""||*sendAmount<0{
      sendCmd.Usage()
      os.Exit(1)
    }
    cli.send(*sendFrom,*sendTo,*sendAmount, nodeid, *sendMine)
  }
  if startnodeCmd.Parsed(){
    nodeid := os.Getenv("NODE_ID")
    if nodeid==""{
      startnodeCmd.Usage()
      os.Exit(1)
    }
    /*
    nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID, *startNodeMiner)
    */
    cli.startnode(nodeid, *startNodeMiner)
  }
}




func (cli *CLI) startnode(nodeid, mineraddress string) {
  fmt.Printf("starting node %s\n", nodeid)
  if len(mineraddress) >0 {
    if wallet.Validateaddress(mineraddress) {
      fmt.Println("mining is on. Address to receive rewards: ", mineraddress)
    } else {
      log.Panic("Wrong miner address!")
    }
  }
  network.Startserver(nodeid, mineraddress)
}






func (cli *CLI) reindexUTXO(nodeid string) {
	bc := bkchain.ContinueBlockchain(nodeid)
  defer bc.Db.Close()
	UTXOSet := bkchain.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.Counttxs()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}




func (cli *CLI) listaddresses(nodeid string){
  wallets,_:=wallet.CreateWallets(nodeid)
  addresses:=wallets.Getalladdress()
  for _, address:=range addresses{
    fmt.Println(address)
  }
}



func(cli *CLI) createwallet(nodeid string){
  wallets,_:=wallet.CreateWallets(nodeid)
  address:=wallets.Addwallet()
  wallets.Savefile(nodeid)
  fmt.Printf("New address is : %s\n", address)
}




func (cli *CLI) printchain(nodeid string){
  bc:=bkchain.ContinueBlockchain(nodeid)
  defer bc.Db.Close()
  bci:=bc.Iterator()

  for {
    block:=bci.Next()
    fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)

    pow:=bkchain.NewProofOfWork(block)
    fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
    for _, tx := range block.Transactions{
      fmt.Println(tx)
    }
    fmt.Println()

    if len(block.PrevBlockHash)==0{
      break
    }
  }
}




func (cli *CLI) createBlockchain(address, nodeid string){
  if !wallet.Validateaddress(address){
    log.Panic("address is not valid")
  }
  chain:=bkchain.Createblockchain(address, nodeid)
  defer chain.Db.Close()

  UTXOSet:=bkchain.UTXOSet{chain}
  UTXOSet.Reindex()
  fmt.Println("Create blockchain")
}




func (cli *CLI) getBalance(address, nodeid string) {
  if !wallet.Validateaddress(address){
    log.Panic("address is not valid")
  }
  bc:=bkchain.ContinueBlockchain(nodeid)
  UTXOSet:=bkchain.UTXOSet{bc}
  defer bc.Db.Close()
  
  balance:=0
  pubkeyhash:=wallet.Base58decode([]byte(address))
  pubkeyhash=pubkeyhash[1:len(pubkeyhash)-4]
  UTXOs := UTXOSet.FindUTXO(pubkeyhash)
  for _, out := range UTXOs{
    balance += out.Value
  }
  fmt.Printf("Balance of '%s': %d\n", address, balance)
}




func (cli *CLI) send(from, to string, amount int, nodeid string, minenow bool){
  if !wallet.Validateaddress(from){
    log.Panic("from is not valid")
  }
  if !wallet.Validateaddress(to){
    log.Panic("to is not valid")
  }
  bc:=bkchain.ContinueBlockchain(nodeid)
  UTXOSet:=bkchain.UTXOSet{bc}
  defer bc.Db.Close()

  wallets, err := wallet.CreateWallets(nodeid)
  if err != nil {
    log.Panic(err)
  }
  wallet := wallets.Getwallet(from)

  tx:=bkchain.NewTransaction(&wallet,to,amount,&UTXOSet)
  if minenow {
    cbtx:=bkchain.Newcoinbasetx(from,"")
    txs:=[]*bkchain.Transaction{cbtx, tx}
    block:=bc.Mineblock(txs)
    UTXOSet.Update(block)
  } else {
    network.Sendtx(network.Knownnodes[0], tx)
    fmt.Println("Send tx")
  }
 fmt.Println("Success!")
}