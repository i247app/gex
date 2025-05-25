package util

import "math/rand"

func GenerateRandomString(length int) string {
	const chars = "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func GenerateSessionKey() string {
	z := GenerateRandomString(4)
	for range 3 {
		z = z + "-" + GenerateRandomString(4)
	}
	return z
}
