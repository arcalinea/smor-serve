package main

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	Pubkey    string `json:"pubkey"`
	CreatedAt uint64 `json:"created_at"`
	Username  string `json:"username"`
}

type SmorServ struct {
	db *leveldb.DB
}

func (ss *SmorServ) forEachItem(user string, f func(*Smor) error) error {
	qrange := &util.Range{
		Start: []byte(user + "/"),
		Limit: []byte(user + "0"), // "0" is the next byte after "/"
	}
	iter := ss.db.NewIterator(qrange, nil)
	defer iter.Release()

	for iter.Next() {
		var s Smor
		if err := json.Unmarshal(iter.Value(), &s); err != nil {
			return err
		}

		if err := f(&s); err != nil {
			return err
		}
	}

	return nil
}

func (ss *SmorServ) getFeed(c echo.Context) error {
	user := c.Param("user")

	// TODO: this is really dumb, i'm just putting everything into a big array, then sending it out.
	// Could instead send each object out as its parsed
	out := []*Smor{}
	err := ss.forEachItem(user, func(s *Smor) error {
		out = append(out, s)
		return nil
	})
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ss *SmorServ) getUser(c echo.Context) error {
	username := c.Param("username")
	fmt.Println("Username", username)

	out := User{}
	data, err := ss.db.Get([]byte(username), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)

	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ss *SmorServ) postFeedItems(c echo.Context) error {
	user := c.Param("user")

	var nudata []Smor
	if err := json.NewDecoder(c.Request().Body).Decode(&nudata); err != nil {
		return err
	}

	b := &leveldb.Batch{}
	for _, sm := range nudata {
		val, err := json.Marshal(sm)
		if err != nil {
			return err
		}

		// TODO: this is using the unix timestamp as the key. This means we will run into issues
		// if two items have the same timestamp. Really, we just want a collection of items, sorted
		// on their timestamp.
		b.Put([]byte(fmt.Sprintf("%s/%d", user, sm.CreatedAt)), val)
	}

	return ss.db.Write(b, nil)
}

func (ss *SmorServ) postNewUser(c echo.Context) error {
	var newUser User
	if err := json.NewDecoder(c.Request().Body).Decode(&newUser); err != nil {
		return err
	}
	fmt.Println("NEW USER", newUser)

	b := &leveldb.Batch{}
	val, err := json.Marshal(newUser)
	if err != nil {
		panic(err)
	}
	fmt.Println("Json val", val)

	// TODO: this is using the username as the key
	b.Put([]byte(fmt.Sprintf(newUser.Username)), val)

	return ss.db.Write(b, nil)
}

func main() {
	db, err := leveldb.OpenFile("smor.db", nil)
	if err != nil {
		panic(err)
	}

	ss := &SmorServ{db: db}

	e := echo.New()
	e.GET("/feed/:user", ss.getFeed)
	e.POST("/feed/:user", ss.postFeedItems)

	e.POST("/user/new", ss.postNewUser)
	e.GET("/user/:username", ss.getUser)

	e.GET("/post/:timestamp", ss.getPost)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Println("ERROR: ", err)
		c.JSON(500, nil)
	}

	panic(e.Start(":7777"))
}
