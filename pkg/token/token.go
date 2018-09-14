package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type Meta struct {
	Account string
}

const (
	expires = 60 * 60 // 一小时过期
)

const (
	reservedPrivateKey = "MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAL4g01D3jJo5p5oTucyXWz1p6F94ouKDwbqNqFtZw4MODACp0N7pFyeVLy96CsWZ6wRoflv0/B89+TnTYDd2yShNW/cz9PCrAxPCpumpRu/C6dSK8Oy9lTrVMbBUxRffDSX+5mbecM/wLde62RF+XJLFSEc/GlIQG4kn90GFVki/AgMBAAECgYBMPY1/YkUXcxcqSc6vo+IKdnWgExf+DSeaT0O7nfswiml1uqLvQDjwvnn1Z9L5+gar9dr1tP+E560Q6xoiI5f1yF9hcf3hObEuHmjN1xgZmjIPaTJnQGKn+isOifbpD7ZsCVMw5j7ERP6GIjZN1mP2mnN7QZKGEj9bjveU0bPkAQJBAN2DF+WQHv+Hyjha4YD4xKUG//5cdgohfg1tK3J0w1BbRvS0YYmWcdgGOGaHjyyvj8XPNLuJMrB6USo69WOWmsECQQDbutn+XcYj8BstBnoXkgnlbdd6j3jHJllzVwVDhG+czh+RfUeO1g6RssZdeI0u+ry4dVqgu/zN2Qm3kWYl30N/AkAwVxF5+Y+qOBn7Xmnj2WYglXx8J/VilJiLmY1ntu+As8qyUEMQ4ZIKkKDyTxcBq3Z2tpdNbc1wEeFwk9lFWHKBAkACky36zR6FTUsEPA8yN4PmLGNaDFReARULRPnK0MJ+E+xKyC0Of3OsQWwRrFf7NPUBNF7bg1hzERgMDqgjyXoBAkAWBWF/PGpZvAZLqVIltjPTs4PhlfuJ2O+ZxHLdg+kLsSp4+9T6fPI/4wm7qXIuZlHUuRugIZk/sUA9QysQ+sEd"
	reservedPublicKey  = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+INNQ94yaOaeaE7nMl1s9aehfeKLig8G6jahbWcODDgwAqdDe6RcnlS8vegrFmesEaH5b9PwfPfk502A3dskoTVv3M/TwqwMTwqbpqUbvwunUivDsvZU61TGwVMUX3w0l/uZm3nDP8C3XutkRflySxUhHPxpSEBuJJ/dBhVZIvwIDAQAB"
)

var keyPair = &struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}{}

func New(account string) (string, error) {
	now := time.Now().Unix()
	claims := &jwt.StandardClaims{
		IssuedAt:  now,
		ExpiresAt: now + expires,
		Issuer:    account,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return t.SignedString(keyPair.PrivateKey)
}

func Parse(raw string) (*Meta, error) {
	c := &jwt.StandardClaims{}
	t, err := jwt.ParseWithClaims(string(raw), c, func(tok *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}

		return keyPair.PublicKey, nil
	})

	if err != nil {
		log.Errorf("verify token error: %s", err.Error())
		return nil, err
	}

	if account := c.Issuer; t.Valid && account != "" {
		return &Meta{Account: account}, nil
	}

	return nil, fmt.Errorf("invalid token: %s", raw)
}

func privateKey(key string) (*rsa.PrivateKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	rsaPriv, err := x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		return nil, fmt.Errorf("illegal rsa private key: %s", key)
	}
	return rsaPriv, nil
}

func publicKey(key string) (*rsa.PublicKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("illegal rsa pulic key: %s", key)
	}
	return rsaPub, nil
}

func init() {
	pubKey, err := publicKey(reservedPublicKey)
	if err != nil {
		panic(err)
	}
	priKey, err := privateKey(reservedPrivateKey)
	if err != nil {
		panic(err)
	}

	keyPair.PrivateKey = priKey
	keyPair.PublicKey = pubKey
}
