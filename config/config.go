package config

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Logger struct {
		Formatter logrus.Formatter
	}
	Application struct {
		Port           string
		Name           string
		AllowedOrigins []string
		Environment    string
		GinMode        string
	}
	BasicAuth struct {
		Username, Password string
	}
	JWT struct {
		PrivateKey, PublicKey []byte
	}
	JWTAdmin struct {
		PrivateKey, PublicKey []byte
	}
	MariaDb struct {
		Driver             string
		Host               string
		Port               string
		Username           string
		Password           string
		Database           string
		DSN                string
		MaxOpenConnections int
		MaxIdleConnections int
	}
}

func (cfg *Config) basicAuth() {
	username := os.Getenv("BASIC_AUTH_USERNAME")
	password := os.Getenv("BASIC_AUTH_PASSWORD")

	cfg.BasicAuth.Username = username
	cfg.BasicAuth.Password = password
}

func (cfg *Config) mariaDb() {
	host := os.Getenv("MARIADB_HOST")
	port := os.Getenv("MARIADB_PORT")
	username := os.Getenv("MARIADB_USERNAME")
	password := os.Getenv("MARIADB_PASSWORD")
	database := os.Getenv("MARIADB_DATABASE")
	maxOpenConnections, _ := strconv.ParseInt(os.Getenv("MARIADB_MAX_OPEN_CONNECTIONS"), 10, 64)
	maxIdleConnections, _ := strconv.ParseInt(os.Getenv("MARIADB_MAX_IDLE_CONNECTIONS"), 10, 64)

	connVal := url.Values{}
	connVal.Add("parseTime", "true")
	connVal.Add("loc", "Asia/Jakarta")

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)
	dsn := fmt.Sprintf("%s?%s", dataSource, connVal.Encode())

	cfg.MariaDb.Driver = "mysql"
	cfg.MariaDb.Host = host
	cfg.MariaDb.Port = port
	cfg.MariaDb.Username = username
	cfg.MariaDb.DSN = dsn
	cfg.MariaDb.Password = password
	cfg.MariaDb.Database = database
	cfg.MariaDb.MaxOpenConnections = int(maxOpenConnections)
	cfg.MariaDb.MaxIdleConnections = int(maxIdleConnections)
}

func (cfg *Config) jwt() {
	privateKey, _ := os.ReadFile("./secret/user/nebeng-dong-private.key")
	publicKey, _ := os.ReadFile("./secret/user/nebeng-dong-public.key")
	// privateKey := []byte(strings.Replace(os.Getenv("JWT_PRIVATE_KEY"), `\n`, "\n", -1))
	// publicKey := []byte(strings.Replace(os.Getenv("JWT_PUBLIC_KEY"), `\n`, "\n", -1))

	cfg.JWT.PrivateKey = privateKey
	cfg.JWT.PublicKey = publicKey
}

func (cfg *Config) jwtAdmin() {
	privateKey, _ := os.ReadFile("./secret/admin/nebeng-dong-admin-private.key")
	publicKey, _ := os.ReadFile("./secret/admin/nebeng-dong-admin-public.key")
	// privateKey := []byte(strings.Replace(os.Getenv("JWT_PRIVATE_KEY"), `\n`, "\n", -1))
	// publicKey := []byte(strings.Replace(os.Getenv("JWT_PUBLIC_KEY"), `\n`, "\n", -1))

	cfg.JWTAdmin.PrivateKey = privateKey
	cfg.JWTAdmin.PublicKey = publicKey
}

func (cfg *Config) app() {
	appName := os.Getenv("APP_NAME")
	port := os.Getenv("PORT")
	environment := os.Getenv("ENVIRONMENT")
	ginMode := os.Getenv("GIN_MODE")
	rawAllowedOrigins := strings.Trim(os.Getenv("ALLOWED_ORIGINS"), " ")

	allowedOrigins := make([]string, 0)
	if rawAllowedOrigins == "" {
		allowedOrigins = append(allowedOrigins, "*")
	} else {
		allowedOrigins = strings.Split(rawAllowedOrigins, ",")
	}

	cfg.Application.Environment = environment
	cfg.Application.GinMode = ginMode
	cfg.Application.Port = port
	cfg.Application.Name = appName
	cfg.Application.AllowedOrigins = allowedOrigins
}
func (cfg *Config) logFormatter() {
	formatter := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			filename := fmt.Sprintf("%s:%d", f.File, f.Line)
			return funcname, filename
		},
	}

	cfg.Logger.Formatter = formatter
}

func Load() *Config {
	cfg := new(Config)
	cfg.app()
	cfg.jwt()
	cfg.jwtAdmin()
	cfg.mariaDb()
	cfg.basicAuth()
	cfg.logFormatter()
	return cfg
}
