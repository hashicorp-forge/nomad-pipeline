package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
)

func loggerMiddleware(logger *zap.Logger, accessLevel string) func(next http.Handler) http.Handler {

	var accessLoggerFn func(msg string, fields ...zap.Field)

	switch accessLevel {
	case zap.DebugLevel.String():
		accessLoggerFn = logger.Debug
	case zap.InfoLevel.String():
		accessLoggerFn = logger.Info
	default:
		panic(fmt.Sprintf("unsupported access log level: %q", accessLevel))
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			startTime := time.Now()

			defer func() {

				// Recover any panicing handler and log the stack trace, so
				// this is available for debugging. Although the output is a
				// little tricky to grok, it is very useful.
				if rec := recover(); rec != nil {
					logger.Error("panic during handling of HTTP request", zap.Reflect("recover_info", rec))
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				accessLoggerFn("successfully handled HTTP request",
					zap.String("remote_address", r.RemoteAddr),
					zap.String("path", r.URL.Path),
					zap.String("proto", r.Proto),
					zap.String("method", r.Method),
					zap.String("user_agent", r.Header.Get("User-Agent")),
					zap.Int("status", ww.Status()),
					zap.Int64("latency_ns", time.Since(startTime).Nanoseconds()),
					zap.Int("content_in_bytes", contentInBytes(r.Header)),
					zap.Int("content_out_bytes", ww.BytesWritten()))
			}()

			next.ServeHTTP(ww, r)

		}
		return http.HandlerFunc(fn)
	}
}

func contentInBytes(header http.Header) int {
	if i, err := strconv.Atoi(header.Get("Content-Length")); err != nil {
		return 0
	} else {
		return i
	}
}

func namespaceCheckMiddleware(stateStore state.State) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ns := getNamespaceParam(r)

			if ns == "" {
				ns = "default"
			}

			// The wildcard namespace glob is not validated here, as it is a
			// valid namespace value for endpoints that support operating
			// across all namespaces but is not a state stored object.
			if ns != "*" {
				if _, err := stateStore.Namespaces().Get(&state.NamespacesGetReq{Name: ns}); err != nil {
					httpWriteResponseError(w, NewResponseError(err.Err(), err.StatusCode()))
					return
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// namespaceWildcardRejectMiddleware rejects any request that uses the wildcard
// namespace ("*"). This is useful for endpoints that do not support operating
// across all namespaces such as create, read, update, and delete operations.
func namespaceWildcardRejectMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if ns := getNamespaceParam(r); ns == "*" {
				httpWriteResponseError(
					w,
					NewResponseError(
						fmt.Errorf("wildcard namespace not allowed here"),
						http.StatusBadRequest),
				)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
