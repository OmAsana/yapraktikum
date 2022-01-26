package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
)

func createSHA256Hash(key string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func Encrypt(data []byte, key string) (string, error) {
	sha256Key := createSHA256Hash(key)
	gcm, err := newGCM(sha256Key)
	if err != nil {
		return "", err
	}

	nonce := nonceFromKey(sha256Key, gcm)
	encrypted := gcm.Seal(nil, nonce, data, nil)

	return string(encrypted), nil
}

func Decrypt(data string, key string) (string, error) {
	sha256Key := createSHA256Hash(key)
	gcm, err := newGCM(sha256Key)
	if err != nil {
		return "", err
	}

	nonce := nonceFromKey(sha256Key, gcm)
	decrypted, err := gcm.Open(nil, nonce, []byte(data), nil)
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
