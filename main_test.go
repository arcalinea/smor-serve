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
  
  user := &User{
    CreatedAt: 0,
    Username: "alice",
  }
  
  if err := ss.saveUser(user); err != nil {
    t.Fatal("Save user failed")
  }
  
  // ml := MerkleList{bs: ss.bs}

  if err := ss.postFeedItems(user.Username, posts); err != nil {
    t.Fatal("Failed postFeedItems")
  }
  fmt.Print("Posted items")
  
  retrievedPosts, err := ss.getFeed(user.Username)
  if err != nil {
    t.Fatal("Failed getFeed")
  }
  
  fmt.Println("Out posts", retrievedPosts)
}