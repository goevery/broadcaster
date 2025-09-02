package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Netflix/go-env"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/juanpmarin/broadcaster/internal/broadcaster"
	"github.com/juanpmarin/broadcaster/internal/handler"
	"github.com/juanpmarin/broadcaster/internal/persistence"
	"github.com/juanpmarin/broadcaster/internal/persistence/mongodb"
	"github.com/juanpmarin/broadcaster/internal/server"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type App struct {
	logger            *zap.Logger
	settings          Settings
	websocketServer   *server.WebSocketServer
	restServer        *server.RESTServer
	persistenceEngine persistence.Engine
}

func NewApp(logger *zap.Logger, settings Settings) *App {
	originChecker := server.NewOriginChecker()
	websocketUpgrader := &websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin:       originChecker.Check,
		EnableCompression: true,
	}

	channelIdValidator := handler.NewChannelIdValidator()

	client, err := mongo.Connect(options.Client().
		ApplyURI(settings.MongoDbUri))
	if err != nil {
		logger.Fatal("failed to connect to mongodb", zap.Error(err))
	}

	persistenceEngine := mongodb.NewPersistenceEngine(client)
	registry := broadcaster.NewInMemoryRegistry(logger)

	heartbeatHandler := handler.NewHeartbeatHandler()
	joinHandler := handler.NewJoinHandler(channelIdValidator, persistenceEngine, registry)
	leaveHandler := handler.NewLeaveHandler(channelIdValidator, registry)
	pushHandler := handler.NewPushHandler(channelIdValidator, persistenceEngine, registry)
	authHandler := handler.NewAuthHandler(settings.JWTSecret)

	router := server.NewRouter(
		logger,
		heartbeatHandler,
		joinHandler,
		leaveHandler,
		pushHandler,
		authHandler,
	)

	websocketServer := server.NewWebSocketServer(
		logger,
		websocketUpgrader,
		registry,
		router,
	)
	restServer := server.NewRESTServer(
		logger,
		pushHandler,
	)

	return &App{
		logger,
		settings,
		websocketServer,
		restServer,
		persistenceEngine,
	}
}

func (a *App) setup(ctx context.Context) error {
	err := a.persistenceEngine.Setup(ctx)
	if err != nil {
		return err
	}

	a.startHttpServer(ctx)

	return nil
}

func (a *App) startHttpServer(ctx context.Context) {
	notifyCtx, notifyCtxCancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCtxCancel()

	address := fmt.Sprintf("0.0.0.0:%d", a.settings.Port)

	router := mux.NewRouter()

	a.websocketServer.Register(router)
	a.restServer.Register(router)

	httpServer := &http.Server{
		Addr:    address,
		Handler: router,
	}

	a.logger.Info("starting http server",
		zap.String("address", address))

	go func() {
		err := httpServer.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Fatal("failed to start http server",
				zap.Error(err))
		}
	}()

	<-notifyCtx.Done()

	a.logger.Info("stopping http server")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCtxCancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		a.logger.Fatal("http server shutdown failed",
			zap.Error(err))
	}

	a.logger.Info("http server stopped")
}

func main() {
	ctx := context.Background()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	var settings Settings
	_, err := env.UnmarshalFromEnviron(&settings)
	if err != nil {
		logger.Fatal("failed to parse settings from environment", zap.Error(err))
	}

	app := NewApp(logger, settings)

	err = app.setup(ctx)
	if err != nil {
		logger.Fatal("failed to setup", zap.Error(err))
	}
}
