package httptemplate

import (
	"net/http"

	chiMiddleware "github.com/go-chi/chi/middleware"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
)

func defaultMiddlewares(app *App) {
	r := app.Router

	r.Use(
		hlog.NewHandler(*app.Logger),
		hlog.URLHandler("url"),
		hlog.MethodHandler("method"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RefererHandler("referer"),
		hlog.RemoteAddrHandler("ip"),
		chiMiddleware.RealIP,
		chiMiddleware.Recoverer,
		chimiddleware.NoCache,
		standardHeaders,
	)
}

func defaultRoutes(app *App) {
	app.Router.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(app.Config.Version))
	})
}

func standardHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}
