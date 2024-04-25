package ginhelper

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IRouter interface {
	Register(r *gin.Engine)
}

type App struct {
	Router     IRouter
	Addr       string
	CtxTimeout time.Duration
	Logger     *zap.Logger
}

type Option func(*App)

func WithRouter(router IRouter) Option {
	return func(app *App) {
		app.Router = router
	}
}

func WithAddr(addr string) Option {
	return func(app *App) {
		app.Addr = addr
	}
}

func WithCtxTimeout(timeout time.Duration) Option {
	return func(app *App) {
		app.CtxTimeout = timeout
	}
}

func NewApp(log *zap.Logger, opts ...Option) *App {
	app := &App{
		Logger:     log,
		Addr:       ":80",
		CtxTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func (app *App) Run(ctx context.Context, middlewares ...gin.HandlerFunc) {
	gin.DisableConsoleColor()
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// trust cloudflare proxy
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{gin.PlatformCloudflare})

	// use logger/recovery middleware, and others system level middlewares
	r.Use(middlewares...)

	// register alive check router
	r.Any("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// register router
	app.Router.Register(r)

	// start server
	svc := &http.Server{
		Addr:    app.Addr,
		Handler: r,
	}
	go func() {
		if err := svc.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("Listen and serve failed", zap.Error(err))
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	c, cancel := context.WithTimeout(ctx, app.CtxTimeout)
	defer cancel()

	// shutdown server
	if err := svc.Shutdown(c); err != nil {
		app.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	// wait for server to shutdown
	<-c.Done()
	app.Logger.Info("timeout", zap.Duration("timeout_of_context", app.CtxTimeout))

	// sync logger
	app.Logger.Sync()
}
