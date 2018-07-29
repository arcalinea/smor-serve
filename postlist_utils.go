package main

import (
	"github.com/ipfs/go-ipfs-blockstore"
)

func (ml *MerkleList) ForEach(f func(*Smor)) error {
	return ml.root.forEach(ml.bs, f)
}

func (mln *MerkleListNode) forEach(bs blockstore.Blockstore, f func(*Smor)) error {
	if len(mln.Posts) > 0 {
		for i := range mln.Posts {
			sm, err := getPost(bs, mln.Posts[i])
			if err != nil {
				return err
			}
			
			f(sm)
		}
	} else if len(mln.Children) > 0 {
		for i := range mln.Children {
			node, err := getNode(bs, mln.Children[i].Node)
			if err != nil {
				return err
			}
			
			if err := node.forEach(bs, f); err != nil {
				return err
			}
		}
	} else {
		// why is it empty 
		panic("merkle node has no posts or children")
	}
	return nil
}