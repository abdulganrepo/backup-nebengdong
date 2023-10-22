package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/Difaal21/nebeng-dong/exception"
	"github.com/Difaal21/nebeng-dong/model"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	BeginTx(ctx context.Context) (tx *sql.Tx, err error)
	RollbackTx(ctx context.Context, tx *sql.Tx) (err error)
	CommitTx(ctx context.Context, tx *sql.Tx) (err error)

	CountFindManyUser(ctx context.Context, params *model.GetManyUserParams) (totalData int64, err error)
	FindManyUser(ctx context.Context, params *model.GetManyUserParams) (users []entity.Users, err error)
	FindOneByEmail(ctx context.Context, email string) (users *entity.Users, err error)
	FindOneById(ctx context.Context, id int64) (users *entity.Users, err error)
	FindOne(ctx context.Context, coloumn string, value any) (user *entity.Users, err error)
	UpdateCoordinate(ctx context.Context, tx *sql.Tx, id int64, coordinate *entity.Coordinate) (err error)
	Insert(ctx context.Context, tx *sql.Tx, users *entity.Users) (id int64, err error)
	Update(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error)
}

type RepositoryImpl struct {
	DB        *sql.DB
	Logger    *logrus.Logger
	TableName string
}

type SqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func NewRepositoryImpl(db *sql.DB, logger *logrus.Logger) Repository {
	return &RepositoryImpl{
		DB:        db,
		Logger:    logger,
		TableName: "users",
	}
}

func (repo *RepositoryImpl) Insert(ctx context.Context, tx *sql.Tx, user *entity.Users) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	INSERT INTO % s
	SET
		id = ?,
		name = ?,
		email = ?,
		phone_number = ?,
		coin = ?,
		coordinate = ?,
		password = ?,
		is_email_verified = ?,
		email_verified_at = ?,
		is_driver = ?,
		created_at = ?,
		updated_at = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, user.ID, user.Name, user.Email, user.PhoneNumber, user.Coin, user.Coordinate, user.Password, user.IsEmailVerified, user.EmailVerifiedAt, user.IsDriver, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
	return
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

func (repo *RepositoryImpl) CountFindManyUser(ctx context.Context, params *model.GetManyUserParams) (totalData int64, err error) {

	var cmd SqlCommand = repo.DB

	q := NewQuery().BaseQueryCountSelectAllUser()

	if params.IsDriver != nil {
		q.AddFilter("u.is_driver", params.IsDriver)
	}

	totalData, err = repo.QueryCount(ctx, cmd, q.GetQuery(), q.GetParams()...)
	if err != nil {
		return
	}

	return
}

func (repo *RepositoryImpl) FindManyUser(ctx context.Context, params *model.GetManyUserParams) (users []entity.Users, err error) {
	var cmd SqlCommand = repo.DB

	var offset = (params.Page - 1) * params.Size

	q := NewQuery().BaseQuerySelectAllUser()

	if params.IsDriver != nil {
		q.AddFilter("u.is_driver", params.IsDriver)
	}

	q.AddLimit(params.Size)
	q.AddOffset(offset)

	users, err = repo.QuerySelectAllUser(ctx, cmd, q.GetQuery(), q.GetParams()...)
	if err != nil {
		return
	}

	lengthOfUsers := len(users)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	return
}

func (repo *RepositoryImpl) FindOne(ctx context.Context, coloumn string, value any) (user *entity.Users, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u.coin,
		ST_X(u.coordinate),
		ST_Y(u.coordinate),
		u.password,
		u.is_email_verified,
		u.email_verified_at,
		u.is_driver,
		u.created_at,
		u.updated_at,
		v.id,
		v.type,
		v.manufacture,
		v.model,
		v.license_plate,
		v.created_at
	FROM 
		%s u
	LEFT JOIN vehicles v ON v.user_id = u.id
	WHERE
		u.%s = ?
	`, repo.TableName, coloumn)

	users, err := repo.Query(ctx, cmd, query, value)
	if err != nil {
		return
	}

	lengthOfUsers := len(users)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	user = &users[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) FindOneByEmail(ctx context.Context, email string) (user *entity.Users, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u.coin,
		ST_X(u.coordinate),
		ST_Y(u.coordinate),
		u.password,
		u.is_email_verified,
		u.email_verified_at,
		u.is_driver,
		u.created_at,
		u.updated_at,
		v.id,
		v.type,
		v.manufacture,
		v.model,
		v.license_plate,
		v.created_at
	FROM 
		%s u
	LEFT JOIN vehicles v ON v.user_id = u.id
	WHERE
		u.email = ?
	`, repo.TableName)

	users, err := repo.Query(ctx, cmd, query, email)
	if err != nil {
		return
	}

	lengthOfUsers := len(users)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	user = &users[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) FindOneById(ctx context.Context, id int64) (user *entity.Users, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u.coin,
		ST_X(u.coordinate),
		ST_Y(u.coordinate),
		u.password,
		u.is_email_verified,
		u.email_verified_at,
		u.is_driver,
		u.created_at,
		u.updated_at,
		v.id,
		v.type,
		v.manufacture,
		v.model,
		v.license_plate,
		v.created_at
	FROM 
		%s u
	LEFT JOIN vehicles v ON v.user_id = u.id
	WHERE
		u.id = ?
	`, repo.TableName)

	users, err := repo.Query(ctx, cmd, query, id)
	if err != nil {
		return
	}

	lengthOfUsers := len(users)
	if lengthOfUsers < 1 {
		err = exception.ErrNotFound
		return
	}

	user = &users[lengthOfUsers-1]

	return
}

func (repo *RepositoryImpl) UpdateCoordinate(ctx context.Context, tx *sql.Tx, id int64, coordinate *entity.Coordinate) (err error) {

	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	UPDATE 
		% s 
	SET 
		coordinate = POINT(?, ?) 
	WHERE 
		id = ?
	`, repo.TableName)

	_, err = Exec(ctx, cmd, command, coordinate.Latitude, coordinate.Longitude, id)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	return
}

func (repo *RepositoryImpl) Update(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error) {
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

func (repo *RepositoryImpl) Query(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (users []entity.Users, err error) {

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

	var user entity.Users
	vehicle := &entity.VehiclesInUser{}
	for rows.Next() {

		var (
			coordinateLatitue   sql.NullFloat64
			coordinateLongitude sql.NullFloat64
		)

		var (
			vehicleId           sql.NullInt64
			vehicleType         sql.NullString
			vehicleManufacture  sql.NullString
			vehicleModel        sql.NullString
			vehicleLicensePlate sql.NullString
			vehicleCreatedAt    sql.NullTime
		)

		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.Coin, &coordinateLatitue, &coordinateLongitude, &user.Password, &user.IsEmailVerified, &user.EmailVerifiedAt, &user.IsDriver, &user.CreatedAt, &user.UpdatedAt, &vehicleId, &vehicleType, &vehicleManufacture, &vehicleModel, &vehicleLicensePlate, &vehicleCreatedAt)

		if err != nil {
			repo.Logger.Error(err.Error())
			return
		}

		if vehicleId.Valid {
			vehicle.ID = vehicleId.Int64
		}

		if vehicleType.Valid {
			vehicle.Type = vehicleType.String
		}

		if vehicleManufacture.Valid {
			vehicle.Manufacture = vehicleManufacture.String
		}

		if vehicleModel.Valid {
			vehicle.Model = vehicleModel.String
		}

		if vehicleLicensePlate.Valid {
			vehicle.LicensePlate = vehicleLicensePlate.String
		}

		if vehicleCreatedAt.Valid {
			vehicle.CreatedAt = vehicleCreatedAt.Time
			user.Vehicles = append(user.Vehicles, vehicle)
		}

		if coordinateLatitue.Valid && coordinateLongitude.Valid {
			user.Coordinate = &entity.Coordinate{
				Latitude:  coordinateLatitue.Float64,
				Longitude: coordinateLongitude.Float64,
			}
		}

		users = append(users, user)
	}

	if users == nil {
		err = exception.ErrNotFound
		return
	}

	return
}

func (repo *RepositoryImpl) QueryCount(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (totalData int64, err error) {
	var rows *sql.Rows

	if rows, err = cmd.QueryContext(ctx, query, args...); err != nil {
		repo.Logger.WithContext(ctx).Error(query, err)
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			repo.Logger.WithContext(ctx).Error(query, err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&totalData)
	}

	return
}

func (repo *RepositoryImpl) QuerySelectAllUser(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (users []entity.Users, err error) {

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

	var user entity.Users
	for rows.Next() {

		var (
			coordinateLatitue   sql.NullFloat64
			coordinateLongitude sql.NullFloat64
		)

		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.PhoneNumber, &user.Coin, &coordinateLatitue, &coordinateLongitude, &user.IsEmailVerified, &user.EmailVerifiedAt, &user.IsDriver, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			repo.Logger.Error(err.Error())
			return
		}

		if coordinateLatitue.Valid && coordinateLongitude.Valid {
			user.Coordinate = &entity.Coordinate{
				Latitude:  coordinateLatitue.Float64,
				Longitude: coordinateLongitude.Float64,
			}
		}

		users = append(users, user)
	}

	if users == nil {
		err = exception.ErrNotFound
		return
	}

	return
}
