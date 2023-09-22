package jwt

import (
	"crypto/rsa"
	"log"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func GetRSAPrivateKey(key []byte) *rsa.PrivateKey {
	signKey, err := jwtv5.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		log.Println("GetRSAPrivateKeyError:", err)
		return nil
	}
	return signKey

}

func GetRSAPublicKey(key []byte) *rsa.PublicKey {
	verifyKey, err := jwtv5.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		log.Println("GetRSAPublicKeyError:", err)
		return nil
	}
	return verifyKey
}
