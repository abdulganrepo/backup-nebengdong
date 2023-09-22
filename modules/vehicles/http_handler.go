package vehicles

import (
	"github.com/Difaal21/nebeng-dong/middleware"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/gin-gonic/gin"
)

var httpResponse = responses.HttpResponseStatusCodesImpl{}

type HTTPHandler struct {
	Usecase Usecase
	Session *middleware.Session
}

func NewHTTPHandler(router *gin.Engine, session *middleware.Session, usecase Usecase) {

	handler := &HTTPHandler{
		Usecase: usecase,
		Session: session,
	}

	router.GET("/nebengdong-service/v1/vehicles", session.Verify, handler.GetAllMyVehicle)
}

func (handler *HTTPHandler) GetAllMyVehicle(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.GetAllMyVehicle(context)
	responses.REST(c, resp)
}
