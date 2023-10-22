package users

import (
	"strings"

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

func NewHTTPHandler(router *gin.Engine, basicAuth *middleware.BasicAuth, session *middleware.Session, usecase Usecase) {

	handler := &HTTPHandler{
		Usecase: usecase,
		Session: session,
	}

	router.POST("/nebengdong-service/v1/users/registration", basicAuth.Verify, handler.Registration)
	// router.POST("/nebengdong-service/v1/users/:id/top-up", basicAuth.Verify, handler.TopUpCoinBalance)
	router.POST("/nebengdong-service/v1/users/login", basicAuth.Verify, handler.Login)
	router.GET("/nebengdong-service/v1/users/profile", session.Verify, handler.GetProfile)
	router.PUT("/nebengdong-service/v1/users/coordinate", session.Verify, handler.UpdateCoordinate)

	router.PUT("/nebengdong-service/v1/users/phone-number", session.Verify, handler.ChangePhoneNumber)
	router.PUT("/nebengdong-service/v1/users/password", session.Verify, handler.ChangePassword)
	router.POST("/nebengdong-service/v1/users/join-driver", session.Verify, handler.JoinAsDriver)
}

func (handler *HTTPHandler) Registration(c *gin.Context) {

	context := c.Request.Context()

	var payload *model.UserRegistration

	if err := c.ShouldBind(&payload); err != nil {
		if errorFields, ok := err.(validator.ValidationErrors); ok {
			schemas := validation.RequestBody(errorFields, payload)
			responses.REST(c, httpResponse.BadRequest("").NewResponses(schemas, "Bad Request"))
			return
		}
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, err.Error()))
		return
	}

	payload.Email = strings.ToLower(payload.Email)
	payload.VehicleLicensePlate = strings.ReplaceAll(payload.VehicleLicensePlate, " ", "")
	resp := handler.Usecase.UserRegistration(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) Login(c *gin.Context) {

	context := c.Request.Context()

	var payload *model.UserLogin

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

	payload.Email = strings.ToLower(payload.Email)
	resp := handler.Usecase.UserLogin(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) GetProfile(c *gin.Context) {
	context := c.Request.Context()

	resp := handler.Usecase.GetProfile(context)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) UpdateCoordinate(c *gin.Context) {
	context := c.Request.Context()

	var payload *model.Coordinate

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

	resp := handler.Usecase.UpdateCoordinate(context, payload)
	responses.REST(c, resp)
}

// func (handler *HTTPHandler) TopUpCoinBalance(c *gin.Context) {

// 	context := c.Request.Context()

// 	userIdStr := c.Param("id")
// 	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

// 	if c.Request.ContentLength < 1 {
// 		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, "request body empty"))
// 		return
// 	}

// 	payload := &model.TopUpCoinBalance{
// 		ID: userId,
// 	}

// 	if err := c.ShouldBind(&payload); err != nil {
// 		if errorFields, ok := err.(validator.ValidationErrors); ok {
// 			schemas := validation.RequestBody(errorFields, payload)
// 			responses.REST(c, httpResponse.BadRequest("").NewResponses(schemas, "Bad Request"))
// 			return
// 		}
// 		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, err.Error()))
// 		return
// 	}

// 	resp := handler.Usecase.TopUpCoinBalance(context, payload)
// 	responses.REST(c, resp)
// }

func (handler *HTTPHandler) ChangePhoneNumber(c *gin.Context) {
	context := c.Request.Context()
	var payload *model.ChangePhoneNumber

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

	resp := handler.Usecase.ChangePhoneNumber(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) ChangePassword(c *gin.Context) {
	context := c.Request.Context()
	var payload *model.ChangePassword

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

	resp := handler.Usecase.ChangePassword(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) JoinAsDriver(c *gin.Context) {
	context := c.Request.Context()
	var payload *model.VehicleRegistration

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

	resp := handler.Usecase.JoinAsDriver(context, payload)
	responses.REST(c, resp)
}
