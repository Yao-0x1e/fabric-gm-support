package aes

import (
	"crypto/cipher"
	"github.com/tjfoc/gmsm/sm4"
)

const BlockSize = sm4.BlockSize

func NewCipher(key []byte) (cipher.Block, error) {
	if len(key) == 24 || len(key) == 32 {
		key = key[:BlockSize]
	}
	return sm4.NewCipher(key)
}
