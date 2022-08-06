package bkchain

import(
  "crypto/sha256"
)

type Merkletree struct {
  Rootnode *Merklenode
}


type Merklenode struct {
  Left *Merklenode
  Right *Merklenode
  Data []byte
}

func Newmerklenode(left, right *Merklenode, data []byte) *Merklenode {
  node:=Merklenode{}

  if left == nil && right == nil {
    hash:= sha256.Sum256(data)
    node.Data=hash[:]
  }else{
    prevhashes:=append(left.Data, right.Data...)
    hash:=sha256.Sum256(prevhashes)
    node.Data=hash[:]
  }
  node.Right=right
  node.Left=left

  return &node
}


func Newmerkletree(data[][]byte) *Merkletree {
  var nodes []Merklenode

  if len(data)%2 != 0 {
    data=append(data, data[len(data)-1])
  }

  for _, dat := range data {
    node:=Newmerklenode(nil,nil,dat)
    nodes=append(nodes,*node)
  }

  for i:=0; i<len(data)/2; i++ {
    var level []Merklenode

    for j:=0; j<len(nodes); j++{
      node:=Newmerklenode(&nodes[j], &nodes[j+1], nil)
      level = append(level, *node)
    }
    nodes=level
  }
  tree:=Merkletree{&nodes[0]}

  return &tree
}

/*

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

}
*/


