package secret

import (
	"crypto/rand"
	b64 "encoding/base64"
	"errors"
)

//Storage of secrets
type Storage struct {
	secrets map[int]string
}

func MakeStorage() *Storage {
	store := new(Storage)
	store.secrets = make(map[int]string)
	store.initSecrets()
	return store
}

func (store Storage) initSecrets() {
	for i := 0; i < 10; i++ {
		b := make([]byte, 10)
		rand.Read(b)
		store.secrets[i] = b64.StdEncoding.EncodeToString(b)
	}
}

func (store Storage) GetSecret(index int) (secret string, err error) {
	if index > len(store.secrets) {
		return "", errors.New("Error. out of bound index")
	}
	return store.secrets[index], nil

}
