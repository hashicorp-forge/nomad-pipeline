package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/go-hclog"
)

func loggerMiddleware(logger hclog.Logger, accessLevel string) func(next http.Handler) http.Handler {

	var accessLoggerFn func(msg string, args ...interface{})

	switch accessLevel {
	case hclog.Trace.String():
		accessLoggerFn = logger.Trace
	case hclog.Debug.String():
		accessLoggerFn = logger.Debug
	case hclog.Info.String():
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
					logger.Error("panic during handling of HTTP request", "recover_info", rec)
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				accessLoggerFn("successfully handled HTTP request",
					"remote_address", r.RemoteAddr,
					"path", r.URL.Path,
					"proto", r.Proto,
					"method", r.Method,
					"user_agent", r.Header.Get("User-Agent"),
					"status", ww.Status(),
					"latency_ns", time.Since(startTime).Nanoseconds(),
					"content_in_bytes", contentInBytes(r.Header),
					"content_out_bytes", ww.BytesWritten())
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
