package mariadb

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Client interface {
	Connect(maxOpenConns int, maxIdleConns int) (db *sql.DB, err error)
	Disconnect(db *sql.DB)
}

type ClientImpl struct {
	DriverName, DataSourceName string
}

func NewClientImpl(driver string, dataSource string) Client {
	return &ClientImpl{DriverName: driver, DataSourceName: dataSource}
}

func (client *ClientImpl) Connect(maxOpenConns int, maxIdleConns int) (db *sql.DB, err error) {
	db, err = sql.Open(client.DriverName, client.DataSourceName)
	if err != nil {
		return
	}

	db.SetMaxIdleConns(maxOpenConns)
	db.SetMaxOpenConns(maxIdleConns)
	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err = db.Ping(); err != nil {
		return
	}
	return
}

func (client *ClientImpl) Disconnect(db *sql.DB) {
	db.Close()
}
