package passengers

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

	Insert(ctx context.Context, tx *sql.Tx, passenger *entity.Passengers) (id int64, err error)
	FindActivePassenger(ctx context.Context, shareRideId int64, userId int64) (passenger *entity.Passengers, err error)
	FindActivePassengerByShareRideId(ctx context.Context, shareRideId int64) (passenger *entity.Passengers, err error)
	UpdateOne(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error)
	FindOnePassengerOnShareRide(ctx context.Context, shareRideId int64, passengerId int64) (passenger *entity.Passengers, err error)
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
		TableName: "passengers",
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

func (repo *RepositoryImpl) Insert(ctx context.Context, tx *sql.Tx, passenger *entity.Passengers) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	INSERT INTO % s
	SET
		id = ?,
		user_id = ?,
		share_ride_id = ?,
		status = ?,
		destination_coordinate = POINT(?, ?),
		distance = ?,
		created_at = ?,
		dropped_at = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, passenger.ID, passenger.UserId, passenger.ShareRideId, passenger.Status, passenger.DestinationCoordinate.Latitude, passenger.DestinationCoordinate.Longitude, passenger.Distance, passenger.CreatedAt, passenger.DroppedAt)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
	return
}

func (repo *RepositoryImpl) FindActivePassenger(ctx context.Context, shareRideId int64, userId int64) (passenger *entity.Passengers, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		p.id,
		p.user_id,
		p.status,
		ST_X (p.destination_coordinate),
		ST_Y (p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		p.share_ride_id
	FROM
		%s p
	WHERE
		p.share_ride_id = ?
		AND p.user_id = ?
		AND p.status IN (1, 2, 3, 4)
	`, repo.TableName)

	passengers, err := repo.Query(ctx, cmd, query, shareRideId, userId)
	if err != nil {
		return
	}

	lengthOfPassengers := len(passengers)
	if lengthOfPassengers < 1 {
		err = exception.ErrNotFound
		return
	}

	passenger = &passengers[lengthOfPassengers-1]

	return
}

func (repo *RepositoryImpl) FindActivePassengerByShareRideId(ctx context.Context, shareRideId int64) (passenger *entity.Passengers, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		p.id,
		p.user_id,
		p.status,
		ST_X (p.destination_coordinate),
		ST_Y (p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		p.share_ride_id
	FROM
		%s p
	WHERE
		p.share_ride_id = ?
		AND p.status IN (1, 2, 3, 4)
	`, repo.TableName)

	passengers, err := repo.Query(ctx, cmd, query, shareRideId)
	if err != nil {
		return
	}

	lengthOfPassengers := len(passengers)
	if lengthOfPassengers < 1 {
		err = exception.ErrNotFound
		return
	}

	passenger = &passengers[lengthOfPassengers-1]

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

func (repo *RepositoryImpl) FindOnePassengerOnShareRide(ctx context.Context, shareRideId int64, passengerId int64) (passenger *entity.Passengers, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		p.id,
		p.user_id,
		p.status,
		ST_X (p.destination_coordinate),
		ST_Y (p.destination_coordinate),
		p.distance,
		p.created_at,
		p.dropped_at,
		p.share_ride_id
	FROM
		%s p
	WHERE
		p.share_ride_id = ? AND p.id = ?
	`, repo.TableName)

	passengers, err := repo.Query(ctx, cmd, query, shareRideId, passengerId)
	if err != nil {
		return
	}

	lengthOfPassengers := len(passengers)
	if lengthOfPassengers < 1 {
		err = exception.ErrNotFound
		return
	}

	passenger = &passengers[lengthOfPassengers-1]

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

func (repo *RepositoryImpl) Query(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (passengers []entity.Passengers, err error) {

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

	var passenger entity.Passengers

	if rows.Next() {
		err = rows.Scan(&passenger.ID, &passenger.UserId, &passenger.Status, &passenger.DestinationCoordinate.Latitude, &passenger.DestinationCoordinate.Longitude, &passenger.Distance, &passenger.CreatedAt, &passenger.DroppedAt, &passenger.ShareRideId)
		if err != nil {
			repo.Logger.Error(err.Error())
			return
		}

		passengers = append(passengers, passenger)
	} else {
		if passengers == nil {
			err = exception.ErrNotFound
			return
		}
	}
	return
}
