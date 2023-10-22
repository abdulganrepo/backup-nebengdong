package administrators

import (
	"strconv"
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

	router.POST("/nebengdong-service/administrators/v1/administrators/login", basicAuth.Verify, handler.Login)
	router.GET("/nebengdong-service/administrators/v1/drivers", session.Verify, handler.GetManyDrivers)
	router.POST("/nebengdong-service/administrators/v1/users/:id/top-up", session.Verify, handler.TopUpCoinBalance)
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
	resp := handler.Usecase.AdminLogin(context, payload)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) GetManyDrivers(c *gin.Context) {
	context := c.Request.Context()

	queryString := c.Request.URL.Query()

	page, _ := strconv.Atoi(queryString.Get("page"))
	size, _ := strconv.Atoi(queryString.Get("size"))

	params := model.GetManyUserParams{
		Page: int64(page),
		Size: int64(size),
	}

	if err := c.ShouldBind(&params); err != nil {
		if errorFields, ok := err.(validator.ValidationErrors); ok {
			schemas := validation.RequestBody(errorFields, params)
			responses.REST(c, httpResponse.BadRequest("").NewResponses(schemas, "Bad Request"))
			return
		}
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, err.Error()))
		return
	}

	resp := handler.Usecase.GetManyDrivers(context, &params)
	responses.REST(c, resp)
}

func (handler *HTTPHandler) TopUpCoinBalance(c *gin.Context) {

	context := c.Request.Context()

	userIdStr := c.Param("id")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)

	if c.Request.ContentLength < 1 {
		responses.REST(c, httpResponse.UnprocessableEntity("").NewResponses(nil, "request body empty"))
		return
	}

	payload := &model.TopUpCoinBalance{
		ID: userId,
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

	resp := handler.Usecase.TopUpCoinBalance(context, payload)
	responses.REST(c, resp)
}
