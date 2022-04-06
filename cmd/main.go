package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	account_handlers "github.com/ProjectOort/oort-server/api/handler/account"
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/account"
	"github.com/ProjectOort/oort-server/conf"
	"github.com/ProjectOort/oort-server/repo"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

func main() {
	cfg := conf.Parse("conf/")
	app := initApp(cfg)
	logger := initLogger(&cfg.Logger)
	cleanup := boostrap(app, logger, cfg)

	done := make(chan struct{})
	go func() {
		err := app.Listen(cfg.Endpoint.HTTP.URL)
		if err != nil {
			logger.Error("[APP] App Error", zap.Error(err))
		}
		done <- struct{}{}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)

	select {
	case <-done:
		return
	case <-sigint:
		err := app.Shutdown()
		logger.Info("[SHUTDOWN] Application start to shutdown...")
		if err != nil {
			logger.Error("[SHUTDOWN] Fiber invoke an error", zap.Error(err))
		}
		cleanup()
	}
}

func initApp(cfg *conf.App) *fiber.App {
	return fiber.New(fiber.Config{
		AppName: cfg.Name,
	})
}

func initLogger(cfg *conf.Logger) *zap.Logger {
	logger, err := zap.NewDevelopment()
	panicIfFailed(err)
	return logger
}

func boostrap(app *fiber.App, logger *zap.Logger, cfg *conf.App) func() {
	// clients
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Repo.Mongo.URL))
	panicIfFailed(err)
	mongoDatabase := mongoClient.Database("oort_server")
	go testMongoConnection(logger, mongoClient)

	// repositories
	accountRepo := repo.NewAccountRepo(mongoDatabase)

	// services
	accountService := account.NewService(logger, &cfg.Biz.Account, accountRepo)

	app.Use(pprof.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	// routes
	api := app.Group("/api/")
	account_handlers.MakeHandlers(api, logger, accountService)

	api.Use(auth.New(logger, accountService))

	return func() {
		printCloseStatus(logger, "Mongo client", mongoClient.Disconnect(context.Background()))
	}
}

func testMongoConnection(log *zap.Logger, _mongo *mongo.Client) {
	time.Sleep(time.Second)
	err := _mongo.Ping(context.Background(), readpref.Primary())
	printConnectStatus(log, "Mongo", err)
}

func panicIfFailed(err error) {
	if err != nil {
		panic(err)
	}
}

func printCloseStatus(log *zap.Logger, name string, err error) {
	log = log.WithOptions(zap.AddCallerSkip(1))
	if err != nil {
		log.Sugar().Errorf("[SHUTDOWN] %s Driver failed to close, error = %+v", name, err)
		return
	}
	log.Sugar().Infof("[SHUTDOWN] %s closed successfully", name)
}

func printConnectStatus(log *zap.Logger, name string, err error) {
	log = log.WithOptions(zap.AddCallerSkip(1))
	if err != nil {
		log.Sugar().Errorf("[INIT] %s connect failed, error = %+v", name, err)
		return
	}
	log.Sugar().Infof("[INIT] Successfully connected to %s", name)
}
