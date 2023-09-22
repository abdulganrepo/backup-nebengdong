package responses

import "github.com/gin-gonic/gin"

func REST(c *gin.Context, r Responses) {
	res := &ResponsesImpl{
		Data:       r.DataProperty(),
		Message:    r.MessageProperty(),
		Status:     r.StatusProperty(),
		Code:       r.CodeProperty(),
		Pagination: r.PaginationProperty(),
	}

	c.Header("Content-Type", "application/json")
	c.JSON(res.Code, r)
	c.Abort()
}
