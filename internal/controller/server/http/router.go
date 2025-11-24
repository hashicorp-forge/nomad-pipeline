package http

import (
	"github.com/go-chi/chi/v5"
)

func newRouter(req *ServerReq) *chi.Mux {

	r := chi.NewRouter()
	r.Use(loggerMiddleware(req.Logger, req.HTPPAccessLogLevel))

	r.Route("/v1", func(r chi.Router) {
		r.Mount("/flows", flowsEndpoint{
			runController: req.Coordinator,
			state:         req.State,
		}.routes())
		r.Mount("/namespaces", namespacesEndpoint{
			state: req.State,
		}.routes())
		r.Mount("/runs", runsEndpoint{
			coordinator: req.Coordinator,
			state:       req.State,
		}.routes())
		r.Mount("/triggers", triggersEndpoint{
			coordinator: req.Coordinator,
			state:       req.State,
		}.routes())
	})

	return r
}
