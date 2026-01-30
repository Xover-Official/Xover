package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type SecurityController struct {
	SecretKey []byte // In production, this would come from Vault/KMS
}

func NewSecurityController(key []byte) *SecurityController {
	return &SecurityController{SecretKey: key}
}

func (sc *SecurityController) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(sc.SecretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (sc *SecurityController) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(sc.SecretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
