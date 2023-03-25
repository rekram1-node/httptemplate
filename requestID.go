package httptemplate

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

const requestIDHeader = "X-Request-Id"

func SendRequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if w.Header().Get(requestIDHeader) == "" {
			w.Header().Add(
				requestIDHeader,
				middleware.GetReqID(ctx),
			)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
