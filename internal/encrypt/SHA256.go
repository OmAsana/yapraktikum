package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func createSHA256Hash(key string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func EncryptSHA256(msg string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

func Encrypt(data []byte, key string) (string, error) {
	sha256Key := createSHA256Hash(key)
	gcm, err := newGCM(sha256Key)
	if err != nil {
		return "", err
	}

	nonce := nonceFromKey(sha256Key, gcm)
	encrypted := gcm.Seal(nil, nonce, data, nil)

	return hex.EncodeToString(encrypted), nil
}

func Decrypt(encrypted string, key string) (string, error) {
	sha256Key := createSHA256Hash(key)
	gcm, err := newGCM(sha256Key)
	if err != nil {
		return "", err
	}

	nonce := nonceFromKey(sha256Key, gcm)
	d, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	decrypted, err := gcm.Open(nil, nonce, d, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), err

}

func nonceFromKey(key []byte, gcm cipher.AEAD) []byte {
	nonce := key[len(key)-gcm.NonceSize():]
	return nonce
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil

}
