package api

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"etym/pkg/algoutil"
	"etym/pkg/log"
	"time"
)

var privateKey *rsa.PrivateKey
var (
	ErrPasswordExpired  = errors.New("expired login infomation")
	ErrInvalidLoginInfo = errors.New("invalid login infomation")
)

type EncryptedPassword struct {
	Password    string `json:"password"`
	EncryptedAt int64  `json:"encrypted_at"`
}

func SetPrivateKeyPath(keyPath string) {
	key, err := algoutil.LoadPrivateKey(keyPath)
	if err != nil {
		panic(err)
	}
	privateKey = key
}

func decryptPassword(cipher string) (string, error) {
	rawPwd, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		log.Error(err.Error())
		return "", ErrInvalidLoginInfo
	}
	dec, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, rawPwd) //RSA解密算法
	if err != nil {
		log.Error(err.Error())
		return "", ErrInvalidLoginInfo

	}
	enc := EncryptedPassword{}
	if err := json.Unmarshal(dec, &enc); err != nil {
		log.Error(err.Error())
		return "", ErrInvalidLoginInfo

	}
	if enc.EncryptedAt < time.Now().Unix()-3600 {
		return "", ErrPasswordExpired
	}

	return enc.Password, nil
}
