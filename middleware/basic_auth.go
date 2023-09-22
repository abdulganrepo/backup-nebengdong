package middleware

import (
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/gin-gonic/gin"
)

var httpResponse = responses.HttpResponseStatusCodesImpl{}

type BasicAuth struct {
	username, password string
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{username, password}
}

func (ba *BasicAuth) Verify(c *gin.Context) {
	username, password, ok := c.Request.BasicAuth()
	if ok && username == ba.username && password == ba.password {
		c.Next()
		return
	}
	responses.REST(c, httpResponse.Unathorized("").NewResponses(nil, "Unauthorized"))
}
