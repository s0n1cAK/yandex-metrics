package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GetHashHex(data []byte, key string) string {
	fmt.Println(key)
	sha := sha256.Sum256(append(data, []byte(key)...))
	return hex.EncodeToString(sha[:])
}
