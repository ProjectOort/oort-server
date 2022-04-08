package main

import (
	"context"
	"fmt"
	"github.com/ProjectOort/oort-server/api/middleware/gerrors"
	"github.com/ProjectOort/oort-server/biz/graph"
	"github.com/ProjectOort/oort-server/biz/search"
	"github.com/olivere/elastic/v7"
	"go.elastic.co/ecszap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	accounthandlers "github.com/ProjectOort/oort-server/api/handler/account"
	asteroidhandlers "github.com/ProjectOort/oort-server/api/handler/asteroid"
	graphhandlers "github.com/ProjectOort/oort-server/api/handler/graph"
	indexhandlers "github.com/ProjectOort/oort-server/api/handler/index"
	searchhandlers "github.com/ProjectOort/oort-server/api/handler/search"
	"github.com/ProjectOort/oort-server/api/middleware/auth"
	"github.com/ProjectOort/oort-server/api/middleware/requestid"
	"github.com/ProjectOort/oort-server/biz/account"
	"github.com/ProjectOort/oort-server/biz/asteroid"
	"github.com/ProjectOort/oort-server/conf"
	"github.com/ProjectOort/oort-server/repo"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

func main() {
	cfg := conf.Parse("conf/")
	logger := initLogger(&cfg.Logger)
	app := initApp(cfg, logger)
	cleanup := boostrap(app, logger, cfg)

	done := make(chan struct{})
	go func() {
		err := app.Listen(cfg.Endpoint.HTTP.URL)
		if err != nil {
			logger.Error("[SHUT_DOWN] App Error", zap.Error(err))
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
		log := logger.Named("[SHUT_DOWN]")
		log.Info("Application start to shutdown...")
		if err != nil {
			log.Error("Fiber occurred an error", zap.Error(err))
		}
		cleanup()
	}
}

func initApp(cfg *conf.App, logger *zap.Logger) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.Name,
		ErrorHandler: gerrors.New(logger),
	})
}

func initLogger(cfg *conf.Logger) *zap.Logger {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	fileEncoderConfig := ecszap.NewDefaultEncoderConfig()

	rotator := &lumberjack.Logger{
		Filename: cfg.Path,
		MaxSize:  cfg.MaxSizeMB,
		MaxAge:   cfg.MaxAgeDay,
		Compress: cfg.Compress,
	}

	var level zapcore.Level
	switch strings.TrimSpace(strings.ToLower(cfg.Level)) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	core := zapcore.NewTee(
		ecszap.WrapCore(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(consoleEncoderConfig),
				zapcore.AddSync(os.Stdout),
				level),
		),
		ecszap.NewCore(
			fileEncoderConfig,
			zapcore.AddSync(rotator),
			level),
	)

	return zap.New(core, zap.AddCaller())
}

func boostrap(app *fiber.App, logger *zap.Logger, cfg *conf.App) func() {
	// clients
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Repo.Mongo.URL))
	panicIfFailed(err)
	mongoDatabase := mongoClient.Database("oort_server")
	go testMongoConnection(logger, mongoClient)

	neo4jDriver, err := neo4j.NewDriver(
		cfg.Repo.Neo4j.URL,
		neo4j.BasicAuth(cfg.Repo.Neo4j.Username, cfg.Repo.Neo4j.Password, cfg.Repo.Neo4j.Realm))
	panicIfFailed(err)
	go testNeo4jConnection(logger, neo4jDriver)

	elasticClient, err := elastic.NewClient(
		elastic.SetURL(cfg.Repo.Elasticsearch.URL),
		elastic.SetBasicAuth(cfg.Repo.Elasticsearch.Username, cfg.Repo.Elasticsearch.Password))
	panicIfFailed(err)
	go testElasticsearchConnection(logger, elasticClient, cfg.Repo.Elasticsearch.URL)

	// repositories
	accountRepo := repo.NewAccountRepo(mongoDatabase)
	asteroidRepo := repo.NewAsteroidRepo(mongoDatabase, neo4jDriver)
	graphRepo := repo.NewGraphRepo(mongoDatabase, neo4jDriver)
	searchRepo := repo.NewSearchRepo(elasticClient)

	// services
	accountService := account.NewService(logger, &cfg.Biz.Account, accountRepo)
	asteroidService := asteroid.NewService(logger, asteroidRepo)
	graphService := graph.NewService(logger, graphRepo)
	searchService := search.NewService(logger, searchRepo)

	app.Use(pprof.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	// routes
	api := app.Group("/api/")
	indexhandlers.RegisterHandlers(api, cfg)
	accounthandlers.RegisterHandlers(api, logger, accountService)

	api.Use(auth.New(logger, accountService))
	asteroidhandlers.RegisterHandlers(api, logger, asteroidService)
	graphhandlers.RegisterHandlers(api, logger, graphService)
	searchhandlers.RegisterHandlers(api, logger, searchService)

	return func() {
		printCloseStatus(logger, "Neo4j driver", neo4jDriver.Close())
		printCloseStatus(logger, "Mongo client", mongoClient.Disconnect(context.Background()))
	}
}

func testMongoConnection(log *zap.Logger, _mongo *mongo.Client) {
	time.Sleep(time.Second)
	err := _mongo.Ping(context.Background(), readpref.Primary())
	printConnectStatus(log, "Mongo", err)
}

func testNeo4jConnection(log *zap.Logger, _neo4j neo4j.Driver) {
	time.Sleep(time.Second)
	err := _neo4j.VerifyConnectivity()
	printConnectStatus(log, "Neo4j", err)
}

func testElasticsearchConnection(log *zap.Logger, _es *elastic.Client, url string) {
	time.Sleep(time.Second)
	_, _, err := _es.Ping(url).Do(context.Background())
	printConnectStatus(log, "Elasticsearch", err)
}

func panicIfFailed(err error) {
	if err != nil {
		panic(err)
	}
}

func printCloseStatus(log *zap.Logger, name string, err error) {
	log = log.WithOptions(zap.AddCallerSkip(1)).Named("[SHUT_DOWN]")
	if err != nil {
		log.Sugar().Errorw(
			fmt.Sprintf("%s Driver failed to close, error:\n%+v", name, err),
			zap.Error(err))
		return
	}
	log.Sugar().Infof("%s closed successfully", name)
}

func printConnectStatus(log *zap.Logger, name string, err error) {
	log = log.WithOptions(zap.AddCallerSkip(1)).Named("[START_UP]")
	if err != nil {
		log.Sugar().Errorw(
			fmt.Sprintf("%s connect failed, error:\n%+v", name, err),
			zap.Error(err))
		return
	}
	log.Sugar().Infof("Successfully connected to %s", name)
}
