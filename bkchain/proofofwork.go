package bkchain

import(
	//"strconv"
  "bytes"
  "encoding/binary"
  "crypto/sha256"
  "math/big"
  "math"
  "log"
)


const targetbits=24

type ProofOfWork struct {
  block *Block
  target *big.Int
}


func NewProofOfWork(b *Block) *ProofOfWork{
  target:=big.NewInt(80)
  target.Lsh(target, uint(256-targetbits))
  pow:=&ProofOfWork{b, target}
  return pow
}


func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Fatal(err)
	}

	return buff.Bytes()
}


func (pow *ProofOfWork) preparedata(nonce int)[]byte {
  data:=bytes.Join(
    [][]byte{
      pow.block.PrevBlockHash,
      pow.block.HashTransactions(),
      IntToHex(pow.block.Timestamp),
      IntToHex(int64(targetbits)),
      IntToHex(int64(nonce)),
    },
    []byte{},
  )
  return data
}


func (pow *ProofOfWork)Run() (int, []byte) {
  var hashInt big.Int
  var hash [32]byte
  nonce:=0

  for nonce<math.MaxInt64 {
    data:=pow.preparedata(nonce)
    hash=sha256.Sum256(data)
    hashInt.SetBytes(hash[:])
    if hashInt.Cmp(pow.target)==-1{
      break
    }else {
      nonce++
    }
  }
  
  return nonce, hash[:]
}


func (pow *ProofOfWork) Validate() bool {
  var hashInt big.Int
  data:=pow.preparedata(pow.block.nonce)
  hash:=sha256.Sum256(data)
  hashInt.SetBytes(hash[:])

  isValid:=hashInt.Cmp(pow.target)==-1
  return isValid
}


