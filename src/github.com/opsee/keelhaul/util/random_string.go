package util

import (
	"math/rand"
)

func RandomByteString(length int, alphabet string) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return b
}

func RandomString(length int, alphabet string) string {
	return string(RandomByteString(length, alphabet)[:length])
}
