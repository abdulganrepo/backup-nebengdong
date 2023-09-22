package payment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	BeginTx(ctx context.Context) (tx *sql.Tx, err error)
	RollbackTx(ctx context.Context, tx *sql.Tx) (err error)
	CommitTx(ctx context.Context, tx *sql.Tx) (err error)

	Insert(ctx context.Context, tx *sql.Tx, payment *entity.Payment) (id int64, err error)
	UpdatePaidStatusByPassengerId(ctx context.Context, tx *sql.Tx, passengerId int64) (err error)
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
		TableName: "payment",
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
func (repo *RepositoryImpl) UpdatePaidStatusByPassengerId(ctx context.Context, tx *sql.Tx, passengerId int64) (err error) {

	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	UPDATE 
		% s 
	SET 
		status = 'paid'
	WHERE 
		passenger_id = ?
	`, repo.TableName)

	_, err = Exec(ctx, cmd, command, passengerId)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	return
}
func (repo *RepositoryImpl) Insert(ctx context.Context, tx *sql.Tx, payment *entity.Payment) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	INSERT INTO % s
	SET
		id = ?,
		passenger_id = ?,
		recipient_id = ?,
		user_id = ?,
		status = ?,
		total_amount = ?,
		created_at = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, payment.ID, payment.PassengerId, payment.RecipientId, payment.UserId, payment.Status, payment.TotalAmount, payment.CreatedAt)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
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
