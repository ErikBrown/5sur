package util

import (
	"crypto/rand"
	"math/big" // rand.Int return a big int
)

// i is the number range
func RandKey(i int64) int64 {
	byteLen := big.NewInt(i)
	// rand.Reader calls an OS function for a random number (much more random
	// than we can generate in Go)
	randNum, err := rand.Int(rand.Reader, byteLen)
	if err != nil {
		panic(err.Error())
	}
	key := randNum.Int64()
	return key
}