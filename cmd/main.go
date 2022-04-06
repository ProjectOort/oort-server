package main

import (
	"context"
	"github.com/ProjectOort/oort-server/biz/graph"
	"github.com/ProjectOort/oort-server/biz/search"
	"github.com/olivere/elastic/v7"
	"os"
	"os/signal"
	"syscall"
	"time"

	account_handlers "github.com/ProjectOort/oort-server/api/handler/account"
	asteroid_handlers "github.com/ProjectOort/oort-server/api/handler/asteroid"
	graph_handlers "github.com/ProjectOort/oort-server/api/handler/graph"
	search_handlers "github.com/ProjectOort/oort-server/api/handler/search"
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
	account_handlers.MakeHandlers(api, logger, accountService)

	api.Use(auth.New(logger, accountService))
	asteroid_handlers.MakeHandlers(api, logger, asteroidService)
	graph_handlers.MakeHandlers(api, logger, graphService)
	search_handlers.MakeHandlers(api, logger, searchService)

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
