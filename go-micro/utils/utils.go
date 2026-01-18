package utils

import (
	"math/rand/v2"
	"strings"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomString(slen int) string {
	var sb strings.Builder
	for range slen {
		num := rand.Int()
		sb.WriteByte(charset[num%len(charset)])
	}
	return sb.String()
}
