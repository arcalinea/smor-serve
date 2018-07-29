package main

import (
	"testing"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-blockstore"
)

func TestBasicPostlist(t *testing.T) {
	// make a memory backed blockstore for testing
	memds := datastore.NewMapDatastore()
	bs := blockstore.NewBlockstore(memds)

	// 'construct' our merklelist
	ml := MerkleList{bs: bs}

	// setup a few random posts to use for test data
	var posts []*Smor
	postLimit := 44
	for i := 1; i <= postLimit; i++ {
		posts = append(posts, getRandomSmor(uint64(i)))
	}

	for _, p := range posts {
		if err := ml.InsertPost(p); err != nil {
			t.Fatal("Failed to split node", err)
		}
		fmt.Println(ml.root)
	}
	
	err := ml.ForEach(func(sm *Smor) error { 
		fmt.Println(sm)
		smor, err := ml.RetrievePost(sm.CreatedAt)
		if err != nil {
			t.Fatal("Error with retrieve func", err)
		}
		fmt.Println("TEST: Found smor: ", smor)
		return nil
	})
	if err != nil {
		t.Fatal("For each func failed")
	}
	
}

