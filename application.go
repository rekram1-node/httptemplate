package httptemplate

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexliesenfeld/health"
	env "github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rekram1-node/httptemplate/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type App struct {
	RootCtx           context.Context
	rootCancelContext context.CancelFunc
	Name              string
	Router            *chi.Mux
	Logger            *zerolog.Logger
	Config            *Configuration
}

type Configuration struct {
	Port     string `env:"PORT" envDefault:"3000"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"DEBUG"`
	Version  string `env:"VERSION" envDefault:"UNKNOWN"`
}

func New(name string) (*App, error) {
	cfg := &Configuration{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	logger := logging.New(
		logging.WithLogLevel(cfg.LogLevel),
		logging.WithServiceName(name),
		logging.WithVersion(cfg.Version),
	)

	ctx := context.Background()
	r := chi.NewRouter()
	ctx, cancelFn := context.WithCancel(logger.WithContext(ctx))

	app := &App{
		Name:              name,
		RootCtx:           ctx,
		rootCancelContext: cancelFn,
		Router:            r,
		Logger:            logger,
		Config:            cfg,
	}

	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		chiMiddleware.RealIP,
		chiMiddleware.Recoverer,
		SendRequestID,
		hlog.NewHandler(*logger),
	)

	defaultMiddlewares(app)
	defaultRoutes(app)

	r.Get("/health", health.NewHandler(health.NewChecker(
		health.WithCacheDuration(1*time.Second),
		health.WithTimeout(10*time.Second),
		health.WithStatusListener(func(ctx context.Context, state health.CheckerState) {
			logger.Info().Msgf("health status changed to %s", state.Status)
		}),
	)))

	return app, nil
}

func (app *App) Start() {
	server := &http.Server{
		Addr:        ":" + app.Config.Port,
		ReadTimeout: 3 * time.Second,
		Handler:     app.Router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal().Err(err).Msg("server closed")
		}
	}()

	app.Logger.Info().Msg("Server Started")
	walkRoutes(app)

	<-done
	app.Logger.Info().Msg("Server Stopped")

	defer func() {
		app.rootCancelContext()
	}()

	if err := server.Shutdown(app.RootCtx); err != nil {
		app.Logger.Fatal().Err(err).Msg("server shutdown failed")
	}
	app.Logger.Info().Msg("Server Exited Properly")
}

func walkRoutes(app *App) {
	_ = chi.Walk(app.Router, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		app.Logger.Info().Str("method", method).Str("path", route).Msg("registered route")
		return nil
	})
}
