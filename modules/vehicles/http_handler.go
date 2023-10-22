package vehicles

import (
	"strconv"

	"github.com/Difaal21/nebeng-dong/helpers/validation"
	"github.com/Difaal21/nebeng-dong/middleware"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	router.PUT("/nebengdong-service/v1/vehicles/:id", session.Verify, handler.UpdateMyVehicle)
}

func (handler *HTTPHandler) GetAllMyVehicle(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.GetAllMyVehicle(context)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) UpdateMyVehicle(c *gin.Context) {
	context := c.Request.Context()
	var payload *model.VehicleRegistration

	vehicleIdStr := c.Param("id")
	vehicleId, _ := strconv.ParseInt(vehicleIdStr, 10, 64)

	if c.Request.ContentLength < 1 {
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, "request body empty"))
		return
	}

	if err := c.ShouldBind(&payload); err != nil {
		if errorFields, ok := err.(validator.ValidationErrors); ok {
			schemas := validation.RequestBody(errorFields, payload)
			responses.REST(c, httpResponse.BadRequest("").NewResponses(schemas, "Bad Request"))
			return
		}
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, err.Error()))
		return
	}

	resp := handler.Usecase.UpdateMyVehicle(context, vehicleId, payload)
	responses.REST(c, resp)
}
