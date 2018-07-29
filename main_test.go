package main

import (
	"testing"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-blockstore"
)

func createSmorServ() *SmorServ {
  db := datastore.NewMapDatastore()
  ss := &SmorServ{db: db, bs: blockstore.NewBlockstore(db)}
  return ss
}

func TestFeed(t *testing.T) {
  ss := createSmorServ()
  
  var posts []*Smor
	postLimit := 11
	for i := 1; i <= postLimit; i++ {
		posts = append(posts, getRandomSmor(uint64(i)))
	}
  
  if err := ss.postFeedItems("alice", posts); err != nil {
    t.Fatal("Failed postFeedItems")
  }
  
  fmt.Println("Posts", posts)
}