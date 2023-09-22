package vehicles

import (
	"context"

	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/sirupsen/logrus"
)

type Usecase interface {
	GetAllMyVehicle(ctx context.Context) responses.Responses
}

type UsecaseImpl struct {
	Repository   Repository
	Logger       *logrus.Logger
	JSONWebToken jwt.JSONWebToken
}

func NewUsecaseImpl(repo Repository, logger *logrus.Logger, jwt jwt.JSONWebToken) Usecase {
	return &UsecaseImpl{
		Repository:   repo,
		Logger:       logger,
		JSONWebToken: jwt,
	}
}

func (u *UsecaseImpl) GetAllMyVehicle(ctx context.Context) responses.Responses {

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	vehicles, err := u.Repository.FindVehiclesByUser(ctx, requester.ID)
	if err != nil && err == exception.ErrInternalServer {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if vehicles == nil {
		return httpResponse.NotFound("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(vehicles, "")
}
