package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp-forge/nomad-pipeline/internal/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

func NewRouter(hclogger hclog.Logger, accessLevel, dataDir string, runC *runner.Controller, state state.State) *chi.Mux {

	httpLogger := hclogger.Named(logger.ComponentNameHTTP)

	r := chi.NewRouter()
	r.Use(loggerMiddleware(httpLogger, accessLevel))

	r.Route("/v1alpha1", func(r chi.Router) {
		r.Mount("/flows", flowsEndpoint{
			runController: runC,
			state:         state,
		}.routes())
		r.Mount("/runs", runsEndpoint{
			dataDir: dataDir,
			state:   state,
		}.routes())
	})

	return r
}
