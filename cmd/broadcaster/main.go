package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/juanpmarin/broadcaster/internal/server"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

var k = koanf.New(".")

type App struct {
	logger          *zap.Logger
	websocketServer *server.WebSocketServer
}

func NewApp(logger *zap.Logger) *App {
	originChecker := server.NewOriginChecker()
	websocketUpgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     originChecker.Check,
	}
	websocketServer := server.NewWebSocketServer(logger, websocketUpgrader)

	return &App{
		logger,
		websocketServer,
	}
}

func (a *App) startHttpServer(ctx context.Context) {
	notifyCtx, notifyCtxCancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer notifyCtxCancel()

	address := fmt.Sprintf("0.0.0.0:%d", 8000)

	mux := http.NewServeMux()
	err := a.websocketServer.Register(mux)
	if err != nil {
		a.logger.Fatal("failed to register websocket server",
			zap.Error(err))
	}

	httpServer := &http.Server{
		Addr:    address,
		Handler: mux,
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

	err = httpServer.Shutdown(shutdownCtx)
	if err != nil {
		a.logger.Fatal("http server shutdown failed",
			zap.Error(err))
	}

	a.logger.Info("http server stopped")
}

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	app := NewApp(logger)

	app.startHttpServer(ctx)
}
