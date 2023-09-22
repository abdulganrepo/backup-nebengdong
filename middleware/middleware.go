package middleware

import (
	"github.com/gin-gonic/gin"
)

type Routes interface {
	Verify(c *gin.Context) any
}
