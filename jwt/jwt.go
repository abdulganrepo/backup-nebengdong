package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken      error = fmt.Errorf("invalid token")
	ErrExpiredOrNotReady error = fmt.Errorf("token is either expired or not ready to use")
)

type JSONWebToken interface {
	CreateToken(ctx context.Context, claims jwtv5.Claims) (tokenString string, err error)
	VerifyToken(ctx context.Context, tokenString string, claims jwtv5.Claims) (err error)
}

type JWT struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func NewJWT(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) JSONWebToken {
	return &JWT{privateKey, publicKey}
}

func (j *JWT) CreateToken(ctx context.Context, claims jwtv5.Claims) (tokenString string, err error) {
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	return token.SignedString(j.PrivateKey)
}

func (j *JWT) VerifyToken(ctx context.Context, tokenString string, claims jwtv5.Claims) (err error) {

	token, err := jwtv5.ParseWithClaims(tokenString, claims, j.keyFunc)
	if err != nil {
		return
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return
}

// The Keyfunc is used by the Parse methods as a callback function to supply the key for verification.
func (j *JWT) keyFunc(token *jwtv5.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwtv5.SigningMethodRSA); !ok {
		return nil, ErrInvalidToken
	}
	return j.PublicKey, nil
}
