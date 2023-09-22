package shareride

import (
	"context"
	"database/sql"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/helpers/date"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/sirupsen/logrus"
)

type PassengerStatusState interface {
	HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses
}

type WaitingStatusHandler struct{}
type PickedUpStatusHandler struct{}
type ArrivedStatusHandler struct{}
type OnTheWayStatusHandler struct{}
type SkippedStatusHandler struct{}
type DoneStatusHandler struct{}

var invalidRule = "INVALID_PASSENGER_STATUS_RULES"

// CurrentStatus: 1 (Menunggu)
func (h *WaitingStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {

	if payload.Code != 2 || payload.Code == -2 {
		return httpResponse.BadRequest(invalidRule).NewResponses(nil, "after waiting should be picked up")
	}

	updatedFieldOnPassenger := map[string]any{
		"status": payload.Code,
	}

	var updatedFieldOnShareRide = make(map[string]any)

	if payload.Code == 2 {
		updatedFieldOnShareRide["is_full"] = true
	}

	if payload.Code == -2 {
		updatedFieldOnShareRide["driver_status"] = 2
		updatedFieldOnShareRide["finisihed_at"] = date.CurrentUTCTime()
	}

	if err := u.Repository.UpdateOne(ctx, tx, shareRide.ID, updatedFieldOnShareRide); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.PassengerRepository.UpdateOne(ctx, tx, payload.ID, updatedFieldOnPassenger); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "status updated")
}

// CurrentStatus: 2 (Diambil atau dijemput)
func (h *PickedUpStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {

	if payload.Code != 3 {
		return httpResponse.BadRequest(invalidRule).NewResponses(nil, "after being picked up the driver should have arrived")
	}

	updatedFieldOnPassenger := map[string]any{
		"status": payload.Code,
	}

	if err := u.PassengerRepository.UpdateOne(ctx, tx, payload.ID, updatedFieldOnPassenger); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	return httpResponse.Ok("").NewResponses(nil, "status updated")
}

// CurrentStatus: 3 (Driver Tiba)
func (h *ArrivedStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {

	if payload.Code != 4 {
		return httpResponse.BadRequest(invalidRule).NewResponses(nil, "once the driver arrives, next is on the way")
	}

	updatedFieldOnPassenger := map[string]any{
		"status": payload.Code,
	}

	if err := u.PassengerRepository.UpdateOne(ctx, tx, payload.ID, updatedFieldOnPassenger); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}
	return httpResponse.Ok("").NewResponses(nil, "status updated")
}

// CurrentStatus: 4 (Dalam Perjalanan)
func (h *OnTheWayStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {

	if payload.Code != 5 {
		return httpResponse.BadRequest(invalidRule).NewResponses(nil, "after on the way it should arrive or done")
	}

	updatedFieldOnShareRide := map[string]any{
		"finished_at":   date.CurrentUTCTime(),
		"driver_status": 2,
	}

	if err := u.Repository.UpdateOne(ctx, tx, shareRide.ID, updatedFieldOnShareRide); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	updatedFieldOnPassenger := map[string]any{
		"status":     payload.Code,
		"dropped_at": date.CurrentUTCTime(),
	}

	if err := u.PassengerRepository.UpdateOne(ctx, tx, payload.ID, updatedFieldOnPassenger); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.PaymentRepository.UpdatePaidStatusByPassengerId(ctx, tx, passenger.ID); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	driver, err := u.UserRepository.FindOneById(ctx, shareRide.DriverId)
	if err != nil && err != exception.ErrNotFound {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger, "driver": driver}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	totalCoin := driver.Coin - passenger.Payment[0].TotalAmount
	reduceBalance := map[string]any{
		"coin": totalCoin,
	}

	if err := u.UserRepository.Update(ctx, tx, shareRide.DriverId, reduceBalance); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}

	if err := u.Repository.CommitTx(ctx, tx); err != nil {
		u.Logger.WithContext(ctx).WithFields(logrus.Fields{"payload": payload, "shareRide": shareRide, "passenger": passenger}).Error(err)
		u.Repository.RollbackTx(ctx, tx)
		return httpResponse.InternalServerError("").NewResponses(nil, err.Error())
	}
	return httpResponse.Ok("").NewResponses(nil, "status updated")
}

// CurrentStatus: -2 (Ditolak)
func (h *SkippedStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {
	return httpResponse.BadRequest(invalidRule).NewResponses(nil, "cannot skip if the passenger is not waiting")
}

// CurrentStatus: 5 (Selesai)
func (h *DoneStatusHandler) HandleUpdateStatus(ctx context.Context, u *UsecaseImpl, payload *model.UpdatePassengerStatus, tx *sql.Tx, shareRide *entity.ShareRide, passenger *entity.Passengers) responses.Responses {
	return httpResponse.BadRequest(invalidRule).NewResponses(nil, "cannot change what has been done")
}
