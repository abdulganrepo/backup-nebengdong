package shareride

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

	router.POST("/nebengdong-service/v1/share-ride/find-passenger", session.Verify, handler.FindPassenger)
	router.POST("/nebengdong-service/v1/share-ride/:id/find-passenger/finish", session.Verify, handler.FinishFindPassenger)
	router.PUT("/nebengdong-service/v1/share-ride/:shareRideId/passenger/:passengerId/status", session.Verify, handler.UpdatePassengerStatusOnShareRide)
	router.GET("/nebengdong-service/v1/share-ride/driver", session.Verify, handler.GetShareRideByDriver)

	router.POST("/nebengdong-service/v1/share-ride/find-driver", session.Verify, handler.FindDriver)
	router.GET("/nebengdong-service/v1/share-ride/passenger", session.Verify, handler.GetShareRideByPassanger)

}

func (handler *HTTPHandler) FindPassenger(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.FindPassenger(context)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) FinishFindPassenger(c *gin.Context) {

	context := c.Request.Context()
	shareRideIdStr := c.Param("id")
	shareRideId, _ := strconv.ParseInt(shareRideIdStr, 10, 64)
	request := model.ShareRideId{
		ID: shareRideId,
	}

	if err := c.ShouldBind(&request); err != nil {
		if errorFields, ok := err.(validator.ValidationErrors); ok {
			schemas := validation.RequestBody(errorFields, shareRideId)
			responses.REST(c, httpResponse.BadRequest("").NewResponses(schemas, "Bad Request"))
			return
		}
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, err.Error()))
		return
	}

	resp := handler.Usecase.FinishFindPassenger(context, shareRideId)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) UpdatePassengerStatusOnShareRide(c *gin.Context) {

	context := c.Request.Context()

	shareRideIdStr := c.Param("shareRideId")
	shareRideId, _ := strconv.ParseInt(shareRideIdStr, 10, 64)
	passengerIdStr := c.Param("passengerId")
	passengerId, _ := strconv.ParseInt(passengerIdStr, 10, 64)

	payload := &model.UpdatePassengerStatus{
		ID:          passengerId,
		ShareRideID: shareRideId,
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

	resp := handler.Usecase.UpdatePassengerStatusOnShareRide(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) FindDriver(c *gin.Context) {

	context := c.Request.Context()

	var payload *model.FindDriver

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

	resp := handler.Usecase.FindDriver(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) GetShareRideByDriver(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.GetShareRideByDriver(context)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) GetShareRideByPassanger(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.GetShareRideByPassanger(context)
	responses.REST(c, resp)
}
