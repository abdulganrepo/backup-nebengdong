package administrators

import (
	"context"
	"database/sql"
	"time"

	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/helpers/cryptography"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/modules/users"
	"github.com/Difaal21/nebeng-dong/responses"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Usecase interface {
	AdminLogin(ctx context.Context, payload *model.UserLogin) responses.Responses
	GetManyDrivers(ctx context.Context, query *model.GetManyUserParams) responses.Responses
	TopUpCoinBalance(ctx context.Context, payload *model.TopUpCoinBalance) responses.Responses
}

type UsecaseImpl struct {
	Logger         *logrus.Logger
	JSONWebToken   jwt.JSONWebToken
	UserRepository users.Repository
}

func NewUsecaseImpl(logger *logrus.Logger, jwt jwt.JSONWebToken, userRepository users.Repository) Usecase {
	return &UsecaseImpl{
		Logger:         logger,
		JSONWebToken:   jwt,
		UserRepository: userRepository,
	}
}

func (u *UsecaseImpl) AdminLogin(ctx context.Context, payload *model.UserLogin) responses.Responses {

	account := map[string]string{
		"name":     "superadmin",
		"email":    "admin@gmail.com",
		"password": "$2a$04$vcdIUZjTQe1MzEom97ySlekUy/vU.vBs.WrDED77YJNuqhNcolMTy",
	}

	if payload.Email != account["email"] {
		return httpResponse.BadRequest("").NewResponses(nil, "invalid credential")
	}

	passwordMatch := cryptography.Verify(account["password"], []byte(payload.Password))
	if !passwordMatch {
		return httpResponse.BadRequest("").NewResponses(nil, "invalid credential")
	}

	expiresAt := time.Now().Add(time.Hour * 7)
	claims := &model.UserBearer{}
	claims.ID = 1
	claims.Email = account["email"]
	claims.Name = account["name"]
	claims.ExpiresAt = jwtv5.NewNumericDate(expiresAt)

	tokenString, err := u.JSONWebToken.CreateToken(ctx, claims)
	if err != nil {
		payload.Password = ""
		u.Logger.WithFields(logrus.Fields{"body": payload}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	result := map[string]any{
		"name":  account["name"],
		"email": account["email"],
		"token": map[string]any{
			"value":     tokenString,
			"expiresAt": expiresAt.Unix(),
		},
	}

	return httpResponse.Ok("").NewResponses(result, "Login success")
}

func (u *UsecaseImpl) GetManyDrivers(ctx context.Context, query *model.GetManyUserParams) responses.Responses {

	var isDriver = true
	query.IsDriver = &isDriver

	totalData, err := u.UserRepository.CountFindManyUser(ctx, query)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	users, err := u.UserRepository.FindManyUser(ctx, query)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if users == nil {
		return httpResponse.NotFound("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponsesOffsetPagination(users, int64(len(users)), totalData, "get many drivers success")
}

func (u *UsecaseImpl) TopUpCoinBalance(ctx context.Context, payload *model.TopUpCoinBalance) responses.Responses {
	var tx *sql.Tx

	user, err := u.UserRepository.FindOneById(ctx, payload.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("requester", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user == nil {
		return httpResponse.NotFound("").NewResponses(nil, err.Error())
	}

	topUpBalance := map[string]any{
		"coin": user.Coin + payload.Coin,
	}

	if err := u.UserRepository.Update(ctx, tx, payload.ID, topUpBalance); err != nil {
		u.Logger.WithField("requester", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "Top up success")
}
