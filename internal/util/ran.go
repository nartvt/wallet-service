package util

import (
	"math/rand"
)

const (
	dictionary = "1234567890"
	codeSize   = 12
)

func generateRandomString(dictionary string, length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = dictionary[rand.Intn(len(dictionary))]
	}

	return string(result)
}

func GenTransactionCode() string {
	return generateRandomString(dictionary, codeSize)
}
