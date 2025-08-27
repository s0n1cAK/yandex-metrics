package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHashHex(data []byte, key string) string {
	sha := sha256.Sum256(append(data, []byte(key)...))
	return hex.EncodeToString(sha[:])
}
