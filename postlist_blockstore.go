package main

import (
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"
)

// Get, Put operations

func putPost(bs blockstore.Blockstore, p *Smor) (*cid.Cid, error) {
	// convert it to cbor ipld
	nd, err := cbor.WrapObject(p, mh.SHA2_256, -1)
	if err != nil {
		return nil, err
	}

	// write it to the blockstore (a content addressing layer over any storage backend)
	if err := bs.Put(nd); err != nil {
		return nil, err
	}

	// return its content identifier
	return nd.Cid(), nil
}

func getPost(bs blockstore.Blockstore, c *cid.Cid) (*Smor, error) {
	// read the data from the datastore
	blk, err := bs.Get(c)
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

func (ml *MerkleListNode) getPostByIndex(bs blockstore.Blockstore, i int) (*Smor, error) {
	return getPost(bs, ml.Posts[i])
}

func putNode(bs blockstore.Blockstore, mln *MerkleListNode) (*cid.Cid, error) {
	node, err := cbor.WrapObject(mln, mh.SHA2_256, -1)
	if err != nil {
		return nil, err
	}
	
	if err := bs.Put(node); err != nil {
		return nil, err
	}
	
	return node.Cid(), nil
}

func getNode(bs blockstore.Blockstore, c *cid.Cid) (*MerkleListNode, error) {
	blk, err := bs.Get(c)
	if err != nil {
		return nil, err
	}
	var mln MerkleListNode
	if err := cbor.DecodeInto(blk.RawData(), &mln); err != nil {
		return nil, err
	}

	return &mln, nil
}
