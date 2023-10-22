package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Difaal21/nebeng-dong/config"
	"github.com/Difaal21/nebeng-dong/databases/mariadb"
	"github.com/Difaal21/nebeng-dong/jwt"
	"github.com/Difaal21/nebeng-dong/middleware"
	"github.com/Difaal21/nebeng-dong/modules/administrators"
	"github.com/Difaal21/nebeng-dong/modules/passengers"
	"github.com/Difaal21/nebeng-dong/modules/payment"
	shareride "github.com/Difaal21/nebeng-dong/modules/share-ride"
	"github.com/Difaal21/nebeng-dong/modules/users"
	"github.com/Difaal21/nebeng-dong/modules/vehicles"
	"github.com/Difaal21/nebeng-dong/responses"
	"github.com/Difaal21/nebeng-dong/server"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload" //for development
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var cfg *config.Config
var httpResponse = responses.HttpResponseStatusCodesImpl{}

func init() {
	cfg = config.Load()
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(cfg.Logger.Formatter)
	logger.SetReportCaller(true)

	basicAuth := middleware.NewBasicAuth(cfg.BasicAuth.Username, cfg.BasicAuth.Password)

	privateKey := jwt.GetRSAPrivateKey(cfg.JWT.PrivateKey)
	publicKey := jwt.GetRSAPublicKey(cfg.JWT.PublicKey)
	jsonWebToken := jwt.NewJWT(privateKey, publicKey)

	session := middleware.NewSession(jsonWebToken)

	privateKeyAdmin := jwt.GetRSAPrivateKey(cfg.JWTAdmin.PrivateKey)
	publicKeyAdmin := jwt.GetRSAPublicKey(cfg.JWTAdmin.PublicKey)
	jsonWebTokenAdmin := jwt.NewJWT(privateKeyAdmin, publicKeyAdmin)
	sessionAdmin := middleware.NewSession(jsonWebTokenAdmin)

	mariaDb := mariadb.NewClientImpl(cfg.MariaDb.Driver, cfg.MariaDb.DSN)
	db, err := mariaDb.Connect(cfg.MariaDb.MaxOpenConnections, cfg.MariaDb.MaxIdleConnections)
	if err != nil {
		logger.Fatal(err)
	}

	gin.SetMode(cfg.Application.GinMode)
	router := gin.New()

	router.GET("/nebengdong-service", index)
	router.NoRoute(notFound)

	vehicleRepository := vehicles.NewRepositoryImpl(db, logger)
	vehicleUsecase := vehicles.NewUsecaseImpl(vehicleRepository, logger, jsonWebToken)
	vehicles.NewHTTPHandler(router, session, vehicleUsecase)

	userRepository := users.NewRepositoryImpl(db, logger)
	userUsecase := users.NewUsecaseImpl(userRepository, logger, vehicleRepository, jsonWebToken)
	users.NewHTTPHandler(router, basicAuth, session, userUsecase)

	adminUsecase := administrators.NewUsecaseImpl(logger, jsonWebTokenAdmin, userRepository)
	administrators.NewHTTPHandler(router, basicAuth, sessionAdmin, adminUsecase)

	passengersRepository := passengers.NewRepositoryImpl(db, logger)

	paymentRepository := payment.NewRepositoryImpl(db, logger)
	paymentDetailRepository := payment.NewPaymentDetailRepositoryImpl(db, logger)

	shareRideRepository := shareride.NewRepositoryImpl(db, logger)
	shareRideUsecase := shareride.NewUsecaseImpl(shareRideRepository, logger, jsonWebToken, passengersRepository, paymentRepository, paymentDetailRepository, userRepository)
	shareride.NewHTTPHandler(router, session, shareRideUsecase)

	handler := cors.New(cors.Options{
		AllowedOrigins:   cfg.Application.AllowedOrigins,
		AllowedMethods:   []string{http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization", "X-RECAPTCHA-TOKEN"},
		AllowCredentials: true,
	}).Handler(router)

	server := server.NewServer(logger, handler, cfg.Application.Port)
	server.Start()

	// When we run this program it will block waiting for a signal. By typing ctrl-C, we can send a SIGINT signal, causing the program to print interrupt and then exit.
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	// closing service for a gracefull shutdown.
	server.Close()
	mariaDb.Disconnect(db)
}

func index(c *gin.Context) {
	responses.REST(c, httpResponse.Ok("").NewResponses(nil, "Ping!!!"))
}

func notFound(c *gin.Context) {
	responses.REST(c, httpResponse.NotFound("").NewResponses(nil, "Page not found!!!"))
}
