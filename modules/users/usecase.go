package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/helpers/cryptography"
	"github.com/Difaal21/nebeng-dong/helpers/date"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/modules/vehicles"
	"github.com/Difaal21/nebeng-dong/responses"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Usecase interface {
	UserRegistration(ctx context.Context, payload *model.UserRegistration) responses.Responses
	UserLogin(ctx context.Context, payload *model.UserLogin) responses.Responses
	GetProfile(ctx context.Context) responses.Responses
	UpdateCoordinate(ctx context.Context, payload *model.Coordinate) responses.Responses
	TopUpCoinBalance(ctx context.Context, payload *model.TopUpCoinBalance) responses.Responses
	ChangePhoneNumber(ctx context.Context, payload *model.ChangePhoneNumber) responses.Responses
	ChangePassword(ctx context.Context, payload *model.ChangePassword) responses.Responses

	JoinAsDriver(ctx context.Context, payload *model.VehicleRegistration) responses.Responses
}

type UsecaseImpl struct {
	Repository        Repository
	Logger            *logrus.Logger
	VehicleRepository vehicles.Repository
	JSONWebToken      jwt.JSONWebToken
}

func NewUsecaseImpl(repo Repository, logger *logrus.Logger, vehicleRepository vehicles.Repository, jwt jwt.JSONWebToken) Usecase {
	return &UsecaseImpl{
		Repository:        repo,
		Logger:            logger,
		VehicleRepository: vehicleRepository,
		JSONWebToken:      jwt,
	}
}

func (u *UsecaseImpl) UserRegistration(ctx context.Context, payload *model.UserRegistration) responses.Responses {
	var tx *sql.Tx
	var err error

	duplicatedEmail, err := u.Repository.FindOneByEmail(ctx, payload.Email)
	if err != nil && err != exception.ErrNotFound {
		payload.Password = ""
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if duplicatedEmail != nil {
		return httpResponse.Conflict("DUPLICATED_EMAIL").NewResponses(nil, "duplicated email when registration")
	}

	duplicatedPhoneNumber, err := u.Repository.FindOne(ctx, "phone_number", payload.PhoneNumber)
	if err != nil && err != exception.ErrNotFound {
		payload.Password = ""
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if duplicatedPhoneNumber != nil {
		return httpResponse.Conflict("DUPLICATED_PHONE_NUMBER").NewResponses(nil, "duplicated phone number when registration")
	}

	hashPassword, err := cryptography.Hash([]byte(payload.Password))
	if err != nil {
		u.Logger.Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, "unexpected error")
	}

	user := entity.Users{
		Name:            payload.Name,
		Email:           payload.Email,
		PhoneNumber:     payload.PhoneNumber,
		Coin:            0,
		Coordinate:      nil,
		Password:        &hashPassword,
		IsDriver:        *payload.IsDriver,
		IsEmailVerified: false,
		EmailVerifiedAt: nil,
		CreatedAt:       *date.CurrentUTCTime(),
		UpdatedAt:       nil,
	}

	if tx, err = u.Repository.BeginTx(ctx); err != nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	userId, err := u.Repository.Insert(ctx, tx, &user)
	if err != nil {
		payload.Password = ""
		u.Logger.WithContext(ctx).WithField("payload", user).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if !(user.IsDriver) {
		if err := u.Repository.CommitTx(ctx, tx); err != nil {
			payload.Password = ""
			u.Logger.WithContext(ctx).WithField("payload", payload).Error(err)
			u.Repository.RollbackTx(ctx, tx)
			return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
		}

		return httpResponse.Created("").NewResponses(nil, "Registration success")
	}

	vehicle := &entity.Vehicle{
		UserId:       userId,
		Type:         "motorcycle",
		Model:        payload.VehicleModel,
		LicensePlate: payload.VehicleLicensePlate,
		Manufacture:  payload.VehicleManufature,
		InUse:        true,
		Capacity:     1,
		CreatedAt:    user.CreatedAt,
	}

	if err := VehicleNullHandler(vehicle); err != nil {
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.BadRequest("").NewResponses(nil, err.Error())
	}

	isVehicleExist, err := u.VehicleRepository.FindOneByLicensePlate(ctx, payload.VehicleLicensePlate)
	if err != nil && err != exception.ErrNotFound {
		payload.Password = ""
		u.Logger.WithField("payload", payload).Error(err.Error())
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if isVehicleExist != nil {
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.Conflict("DUPLICATED_LICENSE_PLATE").NewResponses(nil, "duplicated license plate when registration")
	}

	_, err = u.VehicleRepository.Insert(ctx, tx, vehicle)
	if err != nil {
		payload.Password = ""
		u.Logger.WithContext(ctx).WithField("payload", vehicle).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		payload.Password = ""
		u.Logger.WithContext(ctx).WithField("payload", payload).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Created("").NewResponses(nil, "Registration success")
}

func (u *UsecaseImpl) UserLogin(ctx context.Context, payload *model.UserLogin) responses.Responses {

	user, err := u.Repository.FindOneByEmail(ctx, payload.Email)
	if err != nil && err == exception.ErrInternalServer {
		payload.Password = ""
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user == nil {
		return httpResponse.BadRequest("").NewResponses(nil, "invalid credential")
	}

	passwordMatch := cryptography.Verify(*user.Password, []byte(payload.Password))
	if !passwordMatch {
		u.Logger.WithFields(logrus.Fields{"body": payload}).Error("invalid credential")
		return httpResponse.BadRequest("").NewResponses(nil, "invalid credential")
	}

	expiresAt := time.Now().Add(time.Hour * 7)
	claims := &model.UserBearer{}
	claims.ID = user.ID
	claims.Email = user.Email
	claims.Name = user.Name
	claims.IsDriver = user.IsDriver
	claims.ExpiresAt = jwtv5.NewNumericDate(expiresAt)

	tokenString, err := u.JSONWebToken.CreateToken(ctx, claims)
	if err != nil {
		payload.Password = ""
		u.Logger.WithFields(logrus.Fields{"body": payload}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	result := UserLogin{
		Name:     user.Name,
		Email:    user.Email,
		IsDriver: user.IsDriver,
		Token: Token{
			Value:     &tokenString,
			ExpiresIn: 60 * 60,
		},
	}

	return httpResponse.Ok("").NewResponses(result, "Login success")
}

func (u *UsecaseImpl) GetProfile(ctx context.Context) responses.Responses {

	requester, err := model.GetRequester(ctx)
	if err != nil {
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	user, err := u.Repository.FindOneById(ctx, requester.ID)
	if err != nil && err == exception.ErrInternalServer {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user == nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.NotFound("").NewResponses(nil, "User not found")
	}

	user.Password = nil
	return httpResponse.Ok("").NewResponses(user, "Detail profile")
}

func (u *UsecaseImpl) UpdateCoordinate(ctx context.Context, payload *model.Coordinate) responses.Responses {
	var tx *sql.Tx

	requester, err := model.GetRequester(ctx)
	if err != nil {
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	coordinate := &entity.Coordinate{
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
	}

	err = u.Repository.UpdateCoordinate(ctx, tx, requester.ID, coordinate)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "Coordinate updated")
}

func (u *UsecaseImpl) TopUpCoinBalance(ctx context.Context, payload *model.TopUpCoinBalance) responses.Responses {
	var tx *sql.Tx

	user, err := u.Repository.FindOneById(ctx, payload.ID)
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

	if err := u.Repository.Update(ctx, tx, payload.ID, topUpBalance); err != nil {
		u.Logger.WithField("requester", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "Top up success")
}

func (u *UsecaseImpl) ChangePhoneNumber(ctx context.Context, payload *model.ChangePhoneNumber) responses.Responses {
	var tx *sql.Tx

	requester, err := model.GetRequester(ctx)
	if err != nil {
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	user, err := u.Repository.FindOne(ctx, "phone_number", payload.New)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("user", user).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user != nil {
		return httpResponse.Conflict("DUPLICATED_PHONE_NUMBER").NewResponses(nil, "duplicated phone number")
	}

	updatedField := map[string]any{
		"phone_number": payload.New,
	}

	if err := u.Repository.Update(ctx, tx, requester.ID, updatedField); err != nil {
		u.Logger.WithField("requester", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "phone number updated successfully")
}

func (u *UsecaseImpl) ChangePassword(ctx context.Context, payload *model.ChangePassword) responses.Responses {
	var tx *sql.Tx

	requester, err := model.GetRequester(ctx)
	if err != nil {
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	user, err := u.Repository.FindOneById(ctx, requester.ID)
	if err != nil && err == exception.ErrInternalServer {
		user.Password = nil
		u.Logger.WithField("user", user).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user == nil {
		return httpResponse.BadRequest("").NewResponses(nil, "invalid credential")
	}

	passwordMatch := cryptography.Verify(*user.Password, []byte(payload.Old))
	if !passwordMatch {
		u.Logger.WithFields(logrus.Fields{"body": payload}).Error("invalid credential")
		return httpResponse.BadRequest("INVALID_CREDENTIAL").NewResponses(nil, "invalid old password")
	}

	hashPassword, err := cryptography.Hash([]byte(payload.New))
	if err != nil {
		u.Logger.Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, "unexpected error")
	}

	updatedField := map[string]any{
		"password": hashPassword,
	}

	if err := u.Repository.Update(ctx, tx, user.ID, updatedField); err != nil {
		u.Logger.WithField("user", user).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "password changed successfully")
}

func (u *UsecaseImpl) JoinAsDriver(ctx context.Context, payload *model.VehicleRegistration) responses.Responses {
	var tx *sql.Tx

	requester, err := model.GetRequester(ctx)
	if err != nil {
		return httpResponse.Forbidden("").NewResponses(nil, err.Error())
	}

	user, err := u.Repository.FindOneById(ctx, requester.ID)
	if err != nil && err == exception.ErrInternalServer {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if user == nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.NotFound("").NewResponses(nil, "User not found")
	}

	if user.IsDriver {
		return httpResponse.Forbidden("").NewResponses(nil, "already a driver")
	}

	isVehicleExist, err := u.VehicleRepository.FindOneByLicensePlate(ctx, payload.VehicleLicensePlate)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if isVehicleExist != nil {
		return httpResponse.Conflict("DUPLICATED_LICENSE_PLATE").NewResponses(nil, "duplicated license plate when registration")
	}

	if tx, err = u.Repository.BeginTx(ctx); err != nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	convertToDriver := map[string]any{
		"is_driver": true,
	}

	if err := u.Repository.Update(ctx, tx, requester.ID, convertToDriver); err != nil {
		u.Repository.RollbackTx(ctx, tx)
		u.Logger.WithField("requester", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	vehicle := &entity.Vehicle{
		UserId:       requester.ID,
		Type:         "motorcycle",
		Model:        payload.VehicleModel,
		LicensePlate: payload.VehicleLicensePlate,
		Manufacture:  payload.VehicleManufature,
		InUse:        true,
		Capacity:     1,
		CreatedAt:    *date.CurrentUTCTime(),
	}

	_, err = u.VehicleRepository.Insert(ctx, tx, vehicle)
	if err != nil {
		u.Logger.WithContext(ctx).WithField("payload", vehicle).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithField("payload", payload).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "")
}
