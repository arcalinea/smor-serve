package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/labstack/echo"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	ds "github.com/ipfs/go-datastore"
	ipfsleveldb "github.com/ipfs/go-ds-leveldb"
)

type Smor struct {
	Type       string      `json:"type"`
	Source     string      `json:"source"`
	Author     string      `json:"author"`
	CreatedAt  uint64      `json:"created_at"`
	Data       interface{} `json:"data"`
	Signature  string      `json:"signature"`
	ResponseTo string      `json:"response_to"`
}

type User struct {
	Pubkey	   string      `json:"pubkey"`
	CreatedAt  uint64			 `json:"created_at"`
	Username 	 string			 `json:"username"`
	PostsRoot *cid.Cid    `json:"posts_root"`     
}

type SmorServ struct {
	db ds.Datastore
	bs blockstore.Blockstore
}

func (ss *SmorServ) forEachItem(username string, f func(*Smor) error) error {
	user := User{}
	key := ds.NewKey(username)
	data, err := ss.db.Get(key)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data.([]byte), &user); err != nil {
		return err
	}
	fmt.Println("post root of user", user.PostsRoot)

	ml, err := LoadMerkleList(ss.bs, user.PostsRoot)
	if err != nil {
		return err
	}
	
	// calls function passed in for every post in user's merkle tree
	return ml.ForEach(f)
}

func (ss *SmorServ) getFeed(username string) ([]*Smor, error) {
	// TODO: this is really dumb, i'm just putting everything into a big array, then sending it out.
	// Could instead send each object out as its parsed
	out := []*Smor{}
	err := ss.forEachItem(username, func(sm *Smor) error {
		out = append(out, sm)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (ss *SmorServ) handleGetFeed(c echo.Context) error {
	username := c.Param("user")
	
  posts, err := ss.getFeed(username)
	if err != nil {
		return err
	}
	
	return c.JSON(200, posts)
}

func (ss *SmorServ) getPost(c echo.Context) error {
	user := c.Param("user")
	timestamp := c.Param("timestamp")
		
	createdAt, err := strconv.Atoi(timestamp)
	if err != nil {
		return err
	}

	out := Smor{}
	key := ds.NewKey(fmt.Sprintf("%s/%d", user, createdAt))
	data, err := ss.db.Get(key)
	if err != nil {
		return err
	}
	fmt.Println(data)
	
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		return err
	}

	return c.JSON(200, out)	
}

func (ss *SmorServ) deletePost(c echo.Context) error {
	user := c.Param("user")
	timestamp := c.Param("timestamp")
		
	createdAt, err := strconv.Atoi(timestamp)
	if err != nil {
		return err
	}
	
	key := ds.NewKey(fmt.Sprintf("%s/%d", user, createdAt))
	if err := ss.db.Delete(key); err != nil {
		return err
	}
	
	return c.JSON(200, nil)
}

func (ss *SmorServ) getUser(c echo.Context) error {
	username := c.Param("username")
	fmt.Println("Username", username)
	
	out := User{}
	key := ds.NewKey(username)
	data, err := ss.db.Get(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		return err
	}

	return c.JSON(200, out)	
}

func (ss *SmorServ) postFeedItems(user string, data []*Smor) error {
		for _, sm := range data {
			val, err := json.Marshal(sm)
			if err != nil {
				return err
			}

			// TODO: this is using the unix timestamp as the key. This means we will run into issues
			// if two items have the same timestamp. Really, we just want a collection of items, sorted
			// on their timestamp.
			key := ds.NewKey(fmt.Sprintf("%s/%d", user, sm.CreatedAt))
			ss.db.Put(key, val)
		}
		return nil
}

func (ss *SmorServ) handlePostFeed(c echo.Context) error {
	user := c.Param("user")

	var nudata []*Smor
	if err := json.NewDecoder(c.Request().Body).Decode(&nudata); err != nil {
		return err
	}
	 
	if err := ss.postFeedItems(user, nudata); err != nil {
		return err
	}

	return nil
}

func (ss *SmorServ) postNewUser(c echo.Context) error {
	var newUser User 
	if err := json.NewDecoder(c.Request().Body).Decode(&newUser); err != nil {
		return err
	}
	fmt.Println("NEW USER", newUser)
	
	// b := &leveldb.Batch{}
	val, err := json.Marshal(newUser)
	if err != nil {
		panic(err)
	}
	fmt.Println("Json val", val)

	// TODO: this is using the username as the key
	key := ds.NewKey(newUser.Username)
	return ss.db.Put(key, val)

	// return ss.db.Write(b, nil)
}


func main() {
	// db, err := leveldb.OpenFile("smor.db", nil)
	db, err := ipfsleveldb.NewDatastore("smor.db", nil)
	if err != nil {
		panic(err)
	}

	ss := &SmorServ{db: db, bs: blockstore.NewBlockstore(db)}

	e := echo.New()
	e.GET("/feed/:user", ss.handleGetFeed)
	e.POST("/feed/:user", ss.handlePostFeed)

	e.POST("/user/new", ss.postNewUser)
	e.GET("/user/:username", ss.getUser)
	
	e.GET("/post/:user/:timestamp", ss.getPost)
	e.DELETE("/post/:user/:timestamp", ss.deletePost)
	
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Println("ERROR: ", err)
		c.JSON(500, nil)
	}

	panic(e.Start(":7777"))
}
