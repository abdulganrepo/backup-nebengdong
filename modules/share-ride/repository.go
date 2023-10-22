package shareride

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	BeginTx(ctx context.Context) (tx *sql.Tx, err error)
	RollbackTx(ctx context.Context, tx *sql.Tx) (err error)
	CommitTx(ctx context.Context, tx *sql.Tx) (err error)

	Insert(ctx context.Context, tx *sql.Tx, shareRide *entity.ShareRide) (id int64, err error)
	UpdateOne(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error)
	CheckActiveDriver(ctx context.Context, driverId int64, driverStatus int8) (shareRide *entity.ShareRide, err error)
	FindActiveDriver(ctx context.Context, driverStatus int8) (shareRide *entity.ShareRide, err error)
	FindOne(ctx context.Context, coloumn string, value any) (shareRide *entity.ShareRide, err error)
	FindActiveShareRideByDriver(ctx context.Context, driverId int64) (shareRide *entity.ShareRide, err error)
	FindActiveShareRideByPassenger(ctx context.Context, passengerId int64) (shareRide *entity.ShareRide, err error)
}

type RepositoryImpl struct {
	DB        *sql.DB
	Logger    *logrus.Logger
	TableName string
}

func NewRepositoryImpl(db *sql.DB, logger *logrus.Logger) Repository {
	return &RepositoryImpl{
		DB:        db,
		Logger:    logger,
		TableName: "share_ride",
	}
}

func (repo *RepositoryImpl) BeginTx(ctx context.Context) (tx *sql.Tx, err error) {
	return repo.DB.BeginTx(ctx, nil)
}

func (repo *RepositoryImpl) RollbackTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Rollback()
}

func (repo *RepositoryImpl) CommitTx(ctx context.Context, tx *sql.Tx) (err error) {
	return tx.Commit()
}

func (repo *RepositoryImpl) Insert(ctx context.Context, tx *sql.Tx, shareRide *entity.ShareRide) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}
	command := fmt.Sprintf(`
	INSERT INTO % s
	SET
		id = ?,
		driver_id = ?,
		is_full = ?,
		driver_status = ?,
		created_at = ?,
		finished_at = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, shareRide.ID, shareRide.DriverId, shareRide.IsFull, shareRide.DriverStatus, shareRide.CreatedAt, shareRide.FinishedAt)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
	return
}

func (repo *RepositoryImpl) CheckActiveDriver(ctx context.Context, driverId int64, status int8) (shareRide *entity.ShareRide, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		sr.id,
		sr.driver_id,
		sr.is_full,
		sr.driver_status,
		sr.created_at,
		sr.finished_at,
		p.id,
		p.status,
		ST_X(p.destination_coordinate),
		ST_Y(p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		pymt.id,
		pymt.status,
		pymt.total_amount,
		pymt.created_at,
		pd.id,
		pd.payment_method,
		pd.amount,
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u2.id,
		u2.name,
		u2.email,
		u2.phone_number
	FROM
		%s sr
		left join passengers p on p.share_ride_id = sr.id
		left join payment pymt on pymt.passenger_id = p.id
		left join payment_detail pd on pd.payment_id = pymt.id
		left join users u on u.id = sr.driver_id 
		left join users u2 on u2.id = p.user_id
	WHERE
		sr.driver_id = ? AND sr.driver_status = ?
	`, repo.TableName)

	shareRides, err := repo.Query(ctx, cmd, query, driverId, status)
	if err != nil {
		return
	}

	lengthOfUsers := len(shareRides)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	shareRide = &shareRides[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) FindActiveDriver(ctx context.Context, status int8) (shareRide *entity.ShareRide, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		sr.id,
		sr.driver_id,
		sr.is_full,
		sr.driver_status,
		sr.created_at,
		sr.finished_at,
		p.id,
		p.status,
		ST_X(p.destination_coordinate),
		ST_Y(p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		pymt.id,
		pymt.status,
		pymt.total_amount,
		pymt.created_at,
		pd.id,
		pd.payment_method,
		pd.amount,
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u2.id,
		u2.name,
		u2.email,
		u2.phone_number
	FROM
		%s sr
		left join passengers p on p.share_ride_id = sr.id
		left join payment pymt on pymt.passenger_id = p.id
		left join payment_detail pd on pd.payment_id = pymt.id
		left join users u on u.id = sr.driver_id 
		left join users u2 on u2.id = p.user_id
	WHERE
		sr.driver_status = ?
	LIMIT 1
	`, repo.TableName)

	shareRides, err := repo.Query(ctx, cmd, query, status)
	if err != nil {
		return
	}

	lengthOfUsers := len(shareRides)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	shareRide = &shareRides[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) UpdateOne(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	var (
		placeholders []string
		values       []interface{}
	)

	for field, value := range updateFields {
		placeholders = append(placeholders, field+" = ?")
		values = append(values, value)
	}

	placeholdersStr := strings.Join(placeholders, ", ")

	command := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", repo.TableName, placeholdersStr)
	values = append(values, id)

	_, err = Exec(ctx, cmd, command, values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return exception.ErrNotFound
		}
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				repo.Logger.Error(err.Error())
				return exception.ErrConflict
			}
		}
		repo.Logger.Error(err.Error())
		return exception.ErrInternalServer
	}
	return
}

func (repo *RepositoryImpl) FindOne(ctx context.Context, coloumn string, value any) (shareRide *entity.ShareRide, err error) {
	var cmd SqlCommand = repo.DB
	query := fmt.Sprintf(`
	SELECT
		sr.id,
		sr.driver_id,
		sr.is_full,
		sr.driver_status,
		sr.created_at,
		sr.finished_at,
		p.id,
		p.status,
		ST_X(p.destination_coordinate),
		ST_Y(p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		pymt.id,
		pymt.status,
		pymt.total_amount,
		pymt.created_at,
		pd.id,
		pd.payment_method,
		pd.amount,
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u2.id,
		u2.name,
		u2.email,
		u2.phone_number
	FROM
		%s sr
		left join passengers p on p.share_ride_id = sr.id
		left join payment pymt on pymt.passenger_id = p.id
		left join payment_detail pd on pd.payment_id = pymt.id
		left join users u on u.id = sr.driver_id 
		left join users u2 on u2.id = p.user_id
	WHERE
		sr.%s = ?
	`, repo.TableName, coloumn)

	shareRides, err := repo.Query(ctx, cmd, query, value)
	if err != nil {
		return
	}

	lengthOfUsers := len(shareRides)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	shareRide = &shareRides[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) FindActiveShareRideByDriver(ctx context.Context, driverId int64) (shareRide *entity.ShareRide, err error) {
	var cmd SqlCommand = repo.DB
	query := fmt.Sprintf(`
	SELECT
		sr.id,
		sr.driver_id,
		sr.is_full,
		sr.driver_status,
		sr.created_at,
		sr.finished_at,
		p.id,
		p.status,
		ST_X(p.destination_coordinate),
		ST_Y(p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		pymt.id,
		pymt.status,
		pymt.total_amount,
		pymt.created_at,
		pd.id,
		pd.payment_method,
		pd.amount,
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u2.id,
		u2.name,
		u2.email,
		u2.phone_number
	FROM
		%s sr
		left join passengers p on p.share_ride_id = sr.id
		left join payment pymt on pymt.passenger_id = p.id
		left join payment_detail pd on pd.payment_id = pymt.id
		left join users u on u.id = sr.driver_id 
		left join users u2 on u2.id = p.user_id
	WHERE
		sr.driver_id = ? AND sr.driver_status = 1
	`, repo.TableName)

	shareRides, err := repo.Query(ctx, cmd, query, driverId)
	if err != nil {
		return
	}

	lengthOfUsers := len(shareRides)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	shareRide = &shareRides[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) FindActiveShareRideByPassenger(ctx context.Context, passengerId int64) (shareRide *entity.ShareRide, err error) {
	var cmd SqlCommand = repo.DB
	query := fmt.Sprintf(`
	SELECT
		sr.id,
		sr.driver_id,
		sr.is_full,
		sr.driver_status,
		sr.created_at,
		sr.finished_at,
		p.id,
		p.status,
		ST_X(p.destination_coordinate),
		ST_Y(p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		pymt.id,
		pymt.status,
		pymt.total_amount,
		pymt.created_at,
		pd.id,
		pd.payment_method,
		pd.amount,
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u2.id,
		u2.name,
		u2.email,
		u2.phone_number
	FROM
		%s sr
		left join passengers p on p.share_ride_id = sr.id
		left join payment pymt on pymt.passenger_id = p.id
		left join payment_detail pd on pd.payment_id = pymt.id
		left join users u on u.id = sr.driver_id
		left join users u2 on u2.id = p.user_id
	WHERE
		u2.id = ? AND p.status IN (1, 2, 3, 4)
	`, repo.TableName)

	shareRides, err := repo.Query(ctx, cmd, query, passengerId)
	if err != nil {
		return
	}

	lengthOfUsers := len(shareRides)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	shareRide = &shareRides[lengthOfUsers-1]

	return
}

// ==================================================================================================================== //
type SqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// ==================================================================================================================== //

func Exec(ctx context.Context, cmd SqlCommand, command string, args ...interface{}) (result sql.Result, err error) {
	var stmt *sql.Stmt
	if stmt, err = cmd.PrepareContext(ctx, command); err != nil {
		return
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			return
		}
	}()

	if result, err = stmt.ExecContext(ctx, args...); err != nil {
		return
	}

	return
}

func (repo *RepositoryImpl) Query(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (shareRides []entity.ShareRide, err error) {

	var rows *sql.Rows
	if rows, err = cmd.QueryContext(ctx, query, args...); err != nil {
		repo.Logger.Error(err.Error())
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			repo.Logger.Error(err.Error())
			return
		}
	}()

	var shareRide entity.ShareRide
	var passenger entity.Passengers
	var payment entity.Payment
	var paymentDetails entity.PaymentDetails

	for rows.Next() {

		var (
			shareRideDriverStatus sql.NullInt16
		)

		var (
			passengerId                             sql.NullInt64
			passengerStatus                         sql.NullInt16
			passengerDestinationCoordinateLatitue   sql.NullFloat64
			passengerDestinationCoordinateLongitude sql.NullFloat64
			passengerDistance                       sql.NullFloat64
			passengerCreatedAt                      sql.NullTime
			passengerDroppedAt                      sql.NullTime
		)

		var (
			paymentId          sql.NullInt64
			paymentStatus      sql.NullString
			paymentTotalAmount sql.NullInt64
			paymentCreatedAt   sql.NullTime
		)

		var (
			paymentDetailId            sql.NullInt64
			paymentDetailPaymentMethod sql.NullString
			paymentDetailAmount        sql.NullInt64
		)

		var (
			userId          sql.NullInt64
			userName        sql.NullString
			userEmail       sql.NullString
			userPhoneNumber sql.NullString
		)

		var (
			driverId          sql.NullInt64
			driverName        sql.NullString
			driverEmail       sql.NullString
			driverPhoneNumber sql.NullString
		)

		err = rows.Scan(&shareRide.ID, &shareRide.DriverId, &shareRide.IsFull, &shareRideDriverStatus, &shareRide.CreatedAt, &shareRide.FinishedAt, &passengerId, &passengerStatus, &passengerDestinationCoordinateLatitue, &passengerDestinationCoordinateLongitude, &passengerDistance, &passengerCreatedAt, &passengerDroppedAt, &paymentId, &paymentStatus, &paymentTotalAmount, &paymentCreatedAt, &paymentDetailId, &paymentDetailPaymentMethod, &paymentDetailAmount, &driverId, &driverName, &driverEmail, &driverPhoneNumber, &userId, &userName, &userEmail, &userPhoneNumber)
		if err != nil {
			repo.Logger.Error(err.Error())
			return
		}

		if shareRideDriverStatus.Valid {
			shareRide.DriverStatus = shareRideDriverStatus.Int16
		}

		if passengerId.Valid {
			passenger = entity.Passengers{
				ID:     passengerId.Int64,
				Status: passengerStatus.Int16,
				DestinationCoordinate: entity.Coordinate{
					Latitude:  passengerDestinationCoordinateLatitue.Float64,
					Longitude: passengerDestinationCoordinateLongitude.Float64,
				},
				Distance:  passengerDistance.Float64,
				CreatedAt: passengerCreatedAt.Time,
				User: &entity.UserInVehicle{
					ID:          userId.Int64,
					Name:        userName.String,
					Email:       userEmail.String,
					PhoneNumber: userPhoneNumber.String,
				},
			}
			shareRide.Passengers = append(shareRide.Passengers, &passenger)
		}

		if paymentId.Valid {
			payment = entity.Payment{
				ID:          paymentId.Int64,
				Status:      paymentStatus.String,
				TotalAmount: paymentTotalAmount.Int64,
				CreatedAt:   paymentCreatedAt.Time,
			}

			passenger.Payment = append(passenger.Payment, &payment)
		}

		if paymentDetailId.Valid {
			paymentDetails = entity.PaymentDetails{
				ID:            paymentDetailId.Int64,
				PaymentMethod: paymentDetailPaymentMethod.String,
				Amount:        paymentDetailAmount.Int64,
			}
			payment.PaymentDetails = append(payment.PaymentDetails, paymentDetails)
		}

		if driverId.Valid {
			shareRide.Driver = &entity.UserInVehicle{
				ID:          driverId.Int64,
				Name:        driverName.String,
				Email:       driverEmail.String,
				PhoneNumber: driverPhoneNumber.String,
			}
		}

		shareRides = append(shareRides, shareRide)
	}

	if shareRides == nil {
		err = exception.ErrNotFound
		return
	}

	return
}
