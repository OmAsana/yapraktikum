package encrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	msg := "some random msg"
	key := "someKey"

	encrypted, err := Encrypt([]byte(msg), key)
	assert.NoError(t, err)

	decrypted, err := Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, msg, decrypted)
}
