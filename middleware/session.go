package middleware

import (
	"context"
	"strings"

	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/gin-gonic/gin"
)

type Session struct {
	JSONWebToken jwt.JSONWebToken
}

func NewSession(jwt jwt.JSONWebToken) *Session {
	return &Session{
		JSONWebToken: jwt,
	}
}

func (session *Session) Verify(c *gin.Context) {
	ctx := c.Request.Context()

	authorizationHeader := c.Request.Header.Get("Authorization")
	if authorizationHeader == "" {
		responses.REST(c, httpResponse.Unathorized("").NewResponses(nil, "Empty header"))
		return
	}

	bearerToken := strings.Split(authorizationHeader, " ")
	if len(bearerToken) != 2 {
		responses.REST(c, httpResponse.Unathorized("").NewResponses(nil, "Header format invalid"))
		return
	}

	tokenString := bearerToken[1]

	claims := &model.UserBearer{}
	if err := session.JSONWebToken.VerifyToken(ctx, tokenString, claims); err != nil {
		responses.REST(c, httpResponse.Unathorized("").NewResponses(nil, "Invalid token"))
		return
	}

	ctx = context.WithValue(context.Background(), &model.Identifier{}, claims)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}
