package main 

import (
  "encoding/hex"
  "math/rand"
)

func getRandomSmor(t uint64) *Smor {
	buf := make([]byte, 16)
	rand.Read(buf)
	data := hex.EncodeToString(buf)

	return &Smor{
		CreatedAt: t,
		Data:      data,
	}
}
