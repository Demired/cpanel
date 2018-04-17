package tools

import (
	"math/rand"
	"time"
)

func Rmac() string {
	str := "0123456789abcdef"
	bytes := []byte(str)
	result := []byte("cc:71:")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 8; i++ {
		if i%2 == 0 && i != 0 {
			result = append(result, ':')
		}
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
