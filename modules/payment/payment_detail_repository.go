package payment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Difaal21/nebeng-dong/entity"
	"github.com/sirupsen/logrus"
)

type PaymentDetailRepository interface {
	InsertDetailPayment(ctx context.Context, tx *sql.Tx, payment *entity.PaymentDetails) (id int64, err error)
}

type PaymentDetailRepositoryImpl struct {
	DB        *sql.DB
	Logger    *logrus.Logger
	TableName string
}

func NewPaymentDetailRepositoryImpl(db *sql.DB, logger *logrus.Logger) PaymentDetailRepository {
	return &RepositoryImpl{
		DB:        db,
		Logger:    logger,
		TableName: "payment_detail",
	}
}

func (repo *RepositoryImpl) InsertDetailPayment(ctx context.Context, tx *sql.Tx, payment *entity.PaymentDetails) (id int64, err error) {
	var cmd SqlCommand = repo.DB

	if tx != nil {
		cmd = tx
	}

	command := fmt.Sprintf(`
	INSERT INTO % s
	SET
		id = ?,
		payment_id = ?,
		payment_method = ?,
		amount = ?
	`, repo.TableName)

	result, err := Exec(ctx, cmd, command, payment.ID, payment.PaymentId, payment.PaymentMethod, payment.Amount)
	if err != nil {
		repo.Logger.WithContext(ctx).Error(command, err.Error())
		return
	}

	if id, err = result.LastInsertId(); err != nil {
		return
	}
	return
}
