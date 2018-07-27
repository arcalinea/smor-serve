package main

import (
	"fmt"
	"encoding/json"
	
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"
)

const postsPerNode = 16

type MerkleList struct {
	bs blockstore.Blockstore

	root *MerkleListNode
}

type MerkleListNode struct {
	Posts    []*cid.Cid
	Children []childLink
	Depth    int
}

type childLink struct {
	// Beg is the lowest timestamp on any post in this child
	Beg uint64

	// End is the highest timestamp on any post in this child
	End uint64

	// Node is the hash link to the child object
	Node *cid.Cid
}
// 
// func (ml *MerkleList) RetrievePost(cid *childLink.Node) error {
// 
// }

// InsertPost inserts the given Smor in order into the merklelist
func (ml *MerkleList) InsertPost(p *Smor) error {
	c, err := ml.putPost(p)
	if err != nil {
		return err
	}

	if ml.root == nil {
		// base case of an empty tree, just make a new node with the thing in it
		ml.root = &MerkleListNode{
			Posts: []*cid.Cid{c},
		}

		return nil
	}

	// pass it off to the recursive function (also pass our 'blockstore' so it can persist state)
	extra, err := ml.root.insertPost(ml.bs, p.CreatedAt, c)
	fmt.Println("Got extra back,", extra)
	if err != nil {
		return err
	}

	if extra != nil {
		fmt.Println("HANDLE SPLIT")
		ml.splitNode(ml.bs, extra)
		panic("TODO: handle splitting")
	}

	return nil
}

func (ml *MerkleList) splitNode(bs blockstore.Blockstore, c *cid.Cid) {
	sm, err := ml.getPost(c); 
	fmt.Println("Post from getPost", sm)
	if err != nil {
		panic(err)
	}
	
	val, err := json.Marshal(sm)
	if err != nil {
		panic(err)
	}
	fmt.Println("Val", val)
}

func (ml *MerkleList) putPost(p *Smor) (*cid.Cid, error) {
	// convert it to cbor ipld
	nd, err := cbor.WrapObject(p, mh.SHA2_256, -1)
	if err != nil {
		return nil, err
	}

	// write it to the blockstore (a content addressing layer over any storage backend)
	if err := ml.bs.Put(nd); err != nil {
		return nil, err
	}

	// return its content identifier
	return nd.Cid(), nil
}

func (ml *MerkleListNode) insertPost(bs blockstore.Blockstore, time uint64, c *cid.Cid) (*cid.Cid, error) {
	fmt.Println("C", c)
	// Base case, no child nodes, insert in this node
	if ml.Depth == 0 {

		// iterate from the end to the front, we expect most 'inserts' to be 'append'
		var i int
		for i = len(ml.Posts) - 1; i >= 0; i-- {
			sm, err := ml.getPostByIndex(bs, i)
			if err != nil {
				return nil, err
			}

			if time >= sm.CreatedAt {
				// insert it here!

				// snippet below from golang slice tricks
				ml.Posts = append(ml.Posts[:i], append([]*cid.Cid{c}, ml.Posts[i:]...)...)
				fmt.Println(ml.Posts)
				break
			}
		}

		if i == -1 {
			// if we make it here, our post occurs before every other post, insert it to the front
			ml.Posts = append([]*cid.Cid{c}, ml.Posts...)
		}

		// now check if we need to split
		if len(ml.Posts) > postsPerNode {
			fmt.Println("Splitting node...")
			/* split this node into two
			Go from:
			  [ ............... ]
			To:
			  [ X X ]
			    | |--------|
			    |          |
			  [ .......]  [ .........]
			*/

			extra := ml.Posts[postsPerNode:]
			fmt.Println("EXTRA", extra)
			ml.Posts = ml.Posts[:postsPerNode]
			fmt.Println("ML posts", ml.Posts)

			if len(extra) > 1 {
				panic("don't handle this case yet")
			}
			
			return extra[0], nil
		}

		return nil, nil
	}

	// recursive case, find the child it belongs in
	for i := len(ml.Children) - 1; i >= 0; i-- {
		if time >= ml.Children[i].Beg || i == 0 {
			var extra *cid.Cid
			err := ml.mutateChild(bs, i, func(cmln *MerkleListNode) error {
				ex, err := cmln.insertPost(bs, time, c)
				if err != nil {
					return err
				}

				extra = ex
				return nil
			})
			if err != nil {
				return nil, err
			}

			if extra != nil {
				panic("TODO: handle splitting")
			}

			return nil, nil
		}
	}
	panic("shouldnt ever get here...")
}

// mutateChild loads the given child from its hash, applys the given function
// to it, then rehashes it and updates the link in the children array
func (ml *MerkleListNode) mutateChild(bs blockstore.Blockstore, i int, f func(*MerkleListNode) error) error {
	ch := ml.Children[i]
	blk, err := bs.Get(ch.Node)
	if err != nil {
		return err
	}

	var childNode MerkleListNode
	if err := cbor.DecodeInto(blk.RawData(), &childNode); err != nil {
		return err
	}

	if err := f(&childNode); err != nil {
		return err
	}

	cbnd, err := cbor.WrapObject(childNode, mh.SHA2_256, -1)
	if err != nil {
		return err
	}

	if err := bs.Put(cbnd); err != nil {
		return err
	}

	ml.Children[i].Node = cbnd.Cid()
	return nil
}

func (ml *MerkleList) getPost(c *cid.Cid) (*Smor, error) {
	blk, err := ml.bs.Get(c)
	if err != nil {
		return nil, err
	}
	// unmarshal it into a smor object
	var out Smor
	if err := cbor.DecodeInto(blk.RawData(), &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (ml *MerkleListNode) getPostByIndex(bs blockstore.Blockstore, i int) (*Smor, error) {
	// read the data from the datastore
	blk, err := bs.Get(ml.Posts[i])
	if err != nil {
			return nil, err
	}
	
	// unmarshal it into a smor object
	var smor Smor
	if err := cbor.DecodeInto(blk.RawData(), &smor); err != nil {
		return nil, err
	}

	return &smor, nil
}
