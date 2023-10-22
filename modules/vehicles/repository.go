package vehicles

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
	Insert(ctx context.Context, tx *sql.Tx, vehicle *entity.Vehicle) (id int64, err error)
	Update(ctx context.Context, tx *sql.Tx, id int64, updateFields map[string]any) (err error)
	FindOne(ctx context.Context, coloumn string, value any) (vehicle *vehicleResponses, err error)
	FindOneByLicensePlate(ctx context.Context, licensePlate string) (vehicle *vehicleResponses, err error)
	FindVehiclesByUser(ctx context.Context, userId int64) (vehicles []vehicleResponses, err error)
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
		TableName: "vehicles",
	}
}

type SqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func (repo *RepositoryImpl) Insert(ctx context.Context, tx *sql.Tx, vehicle *entity.Vehicle) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	INSERT INTO %s
	SET
		id = ?,
		user_id = ?,
		type = ?,
		license_plate = ?,
		in_use = ?,
		capacity = ?,
		model = ?,
		manufacture = ?,
		created_at = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, vehicle.ID, vehicle.UserId, vehicle.Type, vehicle.LicensePlate, vehicle.InUse, vehicle.Capacity, vehicle.Model, vehicle.Manufacture, vehicle.CreatedAt)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
	return
}

func (repo *RepositoryImpl) FindOne(ctx context.Context, coloumn string, value any) (vehicle *vehicleResponses, err error) {

	var cmd SqlCommand = repo.DB
	query := fmt.Sprintf(`
	SELECT
		v.id,
		v.type,
		v.model,
		v.license_plate,
		v.manufacture,
		v.in_use,
		v.capacity,
		v.created_at,
		u.id,
		u.name,
		u.email,
		u.phone_number
	FROM %s v
	LEFT JOIN users u ON u.id = v.user_id
	WHERE u.%s = ?
	`, repo.TableName, coloumn)

	vehicles, err := repo.Query(ctx, cmd, query, value)
	if err != nil {
		return
	}

	lengthOfVehicles := len(vehicles)
	if lengthOfVehicles < 1 {
		err = exception.ErrNotFound
		return
	}

	vehicle = &vehicles[lengthOfVehicles-1]

	return

}

func (repo *RepositoryImpl) FindOneByLicensePlate(ctx context.Context, licensePlate string) (vehicle *vehicleResponses, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		v.id,
		v.type,
		v.model,
		v.license_plate,
		v.manufacture,
		v.in_use,
		v.capacity,
		v.created_at,
		u.id,
		u.name,
		u.email,
		u.phone_number
	FROM %s v
	LEFT JOIN users u ON u.id = v.user_id
	Where v.license_plate = ?
	`, repo.TableName)

	vehicles, err := repo.Query(ctx, cmd, query, licensePlate)
	if err != nil {
		return
	}

	lengthOfVehicles := len(vehicles)
	if lengthOfVehicles < 1 {
		err = exception.ErrNotFound
		return
	}

	vehicle = &vehicles[lengthOfVehicles-1]

	return
}

func (repo *RepositoryImpl) FindVehiclesByUser(ctx context.Context, userId int64) (vehicles []vehicleResponses, err error) {
	var cmd SqlCommand = repo.DB

	query := fmt.Sprintf(`
	SELECT
		v.id,
		v.type,
		v.model,
		v.license_plate,
		v.manufacture,
		v.in_use,
		v.capacity,
		v.created_at,
		u.id,
		u.name,
		u.email,
		u.phone_number
	FROM %s v
	LEFT JOIN users u ON u.id = v.user_id
	Where v.user_id = ?
	`, repo.TableName)

	vehicles, err = repo.Query(ctx, cmd, query, userId)
	if err != nil {
		return
	}

	lengthOfVehicles := len(vehicles)
	if lengthOfVehicles < 1 {
		err = exception.ErrNotFound
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

// ====================================================================== //

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

func (repo *RepositoryImpl) Query(ctx context.Context, cmd SqlCommand, query string, args ...interface{}) (vehicles []vehicleResponses, err error) {

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

	var userId sql.NullInt64
	var nameOfUser sql.NullString
	var email sql.NullString
	var phoneNumber sql.NullString

	var vehicle vehicleResponses
	if rows.Next() {
		err = rows.Scan(&vehicle.ID, &vehicle.Type, &vehicle.Model, &vehicle.LicensePlate, &vehicle.Manufacture, &vehicle.InUse, &vehicle.Capacity, &vehicle.CreatedAt, &userId, &nameOfUser, &email, &phoneNumber)
		if err != nil {
			repo.Logger.Error(err.Error())
			return
		}

		if userId.Valid {
			vehicle.Users = &entity.UserInVehicle{
				ID:    userId.Int64,
				Name:  nameOfUser.String,
				Email: email.String,
			}
		}

		vehicles = append(vehicles, vehicle)
	} else {
		if vehicles == nil {
			err = exception.ErrNotFound
			return
		}
	}

	return
}
