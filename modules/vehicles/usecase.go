package vehicles

import (
	"context"
	"database/sql"

	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/sirupsen/logrus"
)

type Usecase interface {
	GetAllMyVehicle(ctx context.Context) responses.Responses
	UpdateMyVehicle(ctx context.Context, id int64, payload *model.VehicleRegistration) responses.Responses
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

func (u *UsecaseImpl) UpdateMyVehicle(ctx context.Context, id int64, payload *model.VehicleRegistration) responses.Responses {

	vehicle, err := u.Repository.FindOne(ctx, "id", id)
	if err != nil && err == exception.ErrInternalServer {
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if vehicle == nil {
		return httpResponse.NotFound("").NewResponses(nil, "Vehicle not found")
	}

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	if requester.ID != vehicle.Users.ID {
		return httpResponse.Forbidden("").NewResponses(nil, "Not eligible to change vehicle data")
	}

	licensePlate, err := u.Repository.FindOneByLicensePlate(ctx, payload.VehicleLicensePlate)
	if err != nil && err == exception.ErrInternalServer {
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if licensePlate != nil {
		return httpResponse.Conflict("LICENSE_PLATE_ALREADY_EXIST").NewResponses(nil, "License plate already exist")
	}

	var tx *sql.Tx
	row := map[string]any{
		"license_plate": payload.VehicleLicensePlate,
		"manufacture":   payload.VehicleManufature,
		"model":         payload.VehicleModel,
	}

	if err := u.Repository.Update(ctx, tx, id, row); err != nil {
		u.Logger.WithFields(logrus.Fields{"vehicleId": id, "payload": payload}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "success to update vehicle data")
}
