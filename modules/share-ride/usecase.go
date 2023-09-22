package shareride

import (
	"context"
	"database/sql"
	"math"
	"os"
	"strconv"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/helpers/date"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/modules/passengers"
	"github.com/Difaal21/nebeng-dong/modules/payment"
	"github.com/Difaal21/nebeng-dong/modules/users"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/sirupsen/logrus"
)

type Usecase interface {
	FindPassenger(ctx context.Context) responses.Responses
	FinishFindPassenger(ctx context.Context, shareRideId int64) responses.Responses
	FindDriver(ctx context.Context, payload *model.FindDriver) responses.Responses
	UpdatePassengerStatusOnShareRide(ctx context.Context, payload *model.UpdatePassengerStatus) responses.Responses
	GetShareRideByDriver(ctx context.Context) responses.Responses
	GetShareRideByPassanger(ctx context.Context) responses.Responses
}

type UsecaseImpl struct {
	Repository              Repository
	Logger                  *logrus.Logger
	JSONWebToken            jwt.JSONWebToken
	PassengerRepository     passengers.Repository
	PaymentRepository       payment.Repository
	PaymentDetailRepository payment.PaymentDetailRepository
	UserRepository          users.Repository
}

func NewUsecaseImpl(repo Repository, logger *logrus.Logger, jwt jwt.JSONWebToken, passengerRepository passengers.Repository, paymentRepository payment.Repository, paymentDetailRepo payment.PaymentDetailRepository, userRepository users.Repository) Usecase {
	return &UsecaseImpl{
		Repository:              repo,
		Logger:                  logger,
		JSONWebToken:            jwt,
		PassengerRepository:     passengerRepository,
		PaymentRepository:       paymentRepository,
		PaymentDetailRepository: paymentDetailRepo,
		UserRepository:          userRepository,
	}
}

func (u *UsecaseImpl) FindPassenger(ctx context.Context) responses.Responses {
	var tx *sql.Tx

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if !requester.IsDriver {
		return httpResponse.Forbidden("NOT_ELIGIBLE").NewResponses(nil, "invalid role")
	}

	minimumBalance, _ := strconv.ParseInt(os.Getenv("MINIMUM_BALANCE"), 10, 64)
	driver, err := u.UserRepository.FindOneById(ctx, requester.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithFields(logrus.Fields{"requester": requester}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if driver.Coin < minimumBalance {
		return httpResponse.Forbidden("").NewResponses(nil, "top up your coin first")
	}

	activeDriver, err := u.Repository.CheckActiveDriver(ctx, requester.ID, 1)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithFields(logrus.Fields{"requester": requester, "activeDriver": activeDriver}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if activeDriver != nil {
		return httpResponse.Conflict("").NewResponses(activeDriver, "please finish previous search")
	}

	shareRide := &entity.ShareRide{
		DriverId:     requester.ID,
		IsFull:       false,
		DriverStatus: 1,
		CreatedAt:    *date.CurrentUTCTime(),
		FinishedAt:   nil,
	}

	_, err = u.Repository.Insert(ctx, tx, shareRide)
	if err != nil {
		u.Logger.WithField("shareRide", shareRide).Error(err.Error())
		u.Logger.WithFields(logrus.Fields{"requester": requester, "activeDriver": activeDriver, "shareRide": shareRide}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, "unexpected error")
	}

	return httpResponse.Created("").NewResponses(nil, "youre active")
}

func (u *UsecaseImpl) FinishFindPassenger(ctx context.Context, shareRideId int64) responses.Responses {

	shareRide, err := u.Repository.FindOne(ctx, "id", shareRideId)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("shareRideId", shareRideId).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if shareRide == nil {
		return httpResponse.NotFound("").NewResponses(nil, "share ride not found")
	}

	if shareRide.DriverStatus == 2 {
		return httpResponse.Conflict("FINISHED_SHARE_RIDE").NewResponses(nil, "share ride already finished")
	}

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if requester.ID != shareRide.DriverId {
		return httpResponse.Forbidden("NOT_ELIGIBLE").NewResponses(nil, "invalid user")
	}

	activePassenger, err := u.PassengerRepository.FindActivePassengerByShareRideId(ctx, shareRide.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithFields(logrus.Fields{"activePassenger": activePassenger, "requester": requester}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if activePassenger != nil {
		return httpResponse.Forbidden("").NewResponses(nil, "Youre share ride still active")
	}

	var tx *sql.Tx
	row := map[string]any{
		"driver_status": 2, // 2 -> DONE
		"finished_at":   date.CurrentUTCTime(),
	}

	if err := u.Repository.UpdateOne(ctx, tx, shareRideId, row); err != nil {
		u.Logger.WithField("shareRideId", shareRideId).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "")
}

func (u *UsecaseImpl) FindDriver(ctx context.Context, payload *model.FindDriver) responses.Responses {

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	activeDriver, err := u.Repository.FindActiveDriver(ctx, 1)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithFields(logrus.Fields{"activeDriver": activeDriver, "requester": requester}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if activeDriver == nil {
		return httpResponse.NotFound("").NewResponses(nil, "driver not found")
	}

	if activeDriver.DriverId == requester.ID {
		return httpResponse.Forbidden("").NewResponses(nil, "youre not allowed to ride with youre self")
	}

	activePassenger, err := u.PassengerRepository.FindActivePassenger(ctx, activeDriver.ID, requester.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithFields(logrus.Fields{"activePassenger": activePassenger, "requester": requester}).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if activePassenger != nil {
		return httpResponse.Conflict("").NewResponses(nil, "Youre share ride still active")
	}

	var tx *sql.Tx

	if tx, err = u.Repository.BeginTx(ctx); err != nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	passenger := &entity.Passengers{
		UserId:      requester.ID,
		ShareRideId: activeDriver.ID,
		Status:      1,
		DestinationCoordinate: entity.Coordinate{
			Latitude:  payload.DestinationCoordinate.Latitude,
			Longitude: payload.DestinationCoordinate.Longitude,
		},
		Distance:  payload.Distance,
		CreatedAt: *date.CurrentUTCTime(),
		DroppedAt: nil,
	}

	passengerId, err := u.PassengerRepository.Insert(ctx, tx, passenger)
	if err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"requester": requester, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	rawTotalAmount := float64(payload.CostPerKM) * payload.Distance
	roundedTotalAmount := int64(math.Round(rawTotalAmount))

	payment := &entity.Payment{
		PassengerId: passengerId,
		RecipientId: activeDriver.DriverId,
		UserId:      requester.ID,
		Status:      "unpaid",
		TotalAmount: roundedTotalAmount,
		CreatedAt:   *date.CurrentUTCTime(),
	}

	paymentId, err := u.PaymentRepository.Insert(ctx, tx, payment)
	if err != nil {
		u.Logger.WithContext(ctx).WithField("payload.payment", payment).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	paymentDetails := &entity.PaymentDetails{
		PaymentId:     paymentId,
		PaymentMethod: "cash",
		Amount:        roundedTotalAmount,
	}

	if _, err = u.PaymentDetailRepository.InsertDetailPayment(ctx, tx, paymentDetails); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload.paymentDetails": paymentDetails, "payload.payment": payment}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithField("payload", payload).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "get driver")
}

func (u *UsecaseImpl) UpdatePassengerStatusOnShareRide(ctx context.Context, payload *model.UpdatePassengerStatus) responses.Responses {

	var tx *sql.Tx

	shareRide, err := u.Repository.FindOne(ctx, "id", payload.ShareRideID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if shareRide == nil {
		return httpResponse.NotFound("").NewResponses(nil, "share ride not found")
	}

	passenger, err := u.PassengerRepository.FindOnePassengerOnShareRide(ctx, shareRide.ID, payload.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("payload", payload).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if passenger == nil {
		return httpResponse.NotFound("").NewResponses(nil, "passenger not found")
	}

	if tx, err = u.Repository.BeginTx(ctx); err != nil {
		u.Logger.WithContext(ctx).Error(err)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	var handler PassengerStatusState

	switch passenger.Status {
	case 1:
		handler = &WaitingStatusHandler{}
	case 2:
		handler = &PickedUpStatusHandler{}
	case 3:
		handler = &ArrivedStatusHandler{}
	case 4:
		handler = &OnTheWayStatusHandler{}
	case 5:
		handler = &DoneStatusHandler{}
	case -2:
		handler = &SkippedStatusHandler{}
	default:
		return httpResponse.BadRequest("INVALID_PASSENGER_STATUS").NewResponses(nil, "invalid passenger status")
	}

	return handler.HandleUpdateStatus(ctx, u, payload, tx, shareRide, passenger)
}

func (u *UsecaseImpl) GetShareRideByDriver(ctx context.Context) responses.Responses {

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if !requester.IsDriver {
		return httpResponse.Forbidden("INVALID_ROLE").NewResponses(nil, "")
	}

	shareRide, err := u.Repository.FindActiveShareRideByDriver(ctx, requester.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if shareRide == nil {
		return httpResponse.NotFound("").NewResponses(nil, "")
	}

	return httpResponse.Ok("").NewResponses(shareRide, "")
}

func (u *UsecaseImpl) GetShareRideByPassanger(ctx context.Context) responses.Responses {

	requester, err := model.GetRequester(ctx)
	if err != nil {
		u.Logger.WithField("requester", requester).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	shareRide, err := u.Repository.FindActiveShareRideByPassenger(ctx, requester.ID)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithField("shareRide", shareRide).Error(err.Error())
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if shareRide == nil {
		return httpResponse.NotFound("").NewResponses(nil, "share ride active not found")
	}

	return httpResponse.Ok("").NewResponses(shareRide, "")
}
