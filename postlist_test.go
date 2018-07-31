package main

import (
	"testing"
	"fmt"
	"math/rand"
	"encoding/json"

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
	postLimit := 55
	var posts []*Smor
	for i := 1; i <= postLimit; i++ {
		posts = append(posts, getRandomSmor(uint64(i*2)))
	}
	
	rand.Shuffle(len(posts), func(i, j int) {
		posts[i], posts[j] = posts[j], posts[i]
	})
	foo, _ := json.Marshal(posts)
	fmt.Println("RANDOM: ", string(foo))
	
	oddSmor := getRandomSmor(7)

	for _, p := range posts {
		if err := ml.InsertPost(p); err != nil {
			t.Fatal("Failed to split node", err)
		}
		fmt.Println(ml.root)
	}
	
	if err := ml.InsertPost(oddSmor); err != nil {
		t.Fatal("Failed to insert post in middle", err)
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

