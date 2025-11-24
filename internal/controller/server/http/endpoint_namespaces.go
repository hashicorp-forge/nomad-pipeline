package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type namespacesEndpoint struct {
	state state.State
}

func (n namespacesEndpoint) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(namespaceWildcardRejectMiddleware())

	router.Route("/", func(r chi.Router) {
		r.Post("/", n.create)
		r.Get("/", n.list)
	})

	router.Route("/{name}", func(r chi.Router) {
		r.Use(n.context)
		r.Delete("/", n.delete)
		r.Get("/", n.get)
	})

	return router
}

type NamespaceCreateReq struct {
	Namespace *sharedstate.Namespace `json:"namespace"`
}

type NamespaceCreateResp struct {
	Namespace            *sharedstate.Namespace `json:"namespace"`
	internalResponseMeta `json:"-"`
}

func (n namespacesEndpoint) create(w http.ResponseWriter, r *http.Request) {

	var req NamespaceCreateReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteResponseError(w, NewResponseError(fmt.Errorf("failed to decode object: %w", err), 400))
		return
	}

	_, err := n.state.Namespaces().Create(&state.NamespacesCreateReq{Namespace: req.Namespace})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := NamespaceCreateResp{
			Namespace:            req.Namespace,
			internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
		}
		httpWriteResponse(w, &resp)
	}
}

type NamespaceDeleteResp struct {
	internalResponseMeta `json:"-"`
}

func (n namespacesEndpoint) delete(w http.ResponseWriter, r *http.Request) {

	ns := r.Context().Value("name").(string)

	if ns == "default" {
		respErr := NewResponseError(
			errors.New("cannot delete default namespace"),
			http.StatusBadRequest,
		)
		httpWriteResponseError(w, respErr)
		return
	}

	_, err := n.state.Namespaces().Delete(&state.NamespacesDeleteReq{Name: ns})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowDeleteResp{
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type NamespaceGetReq struct {
	ID string `json:"id"`
}

type NamespaceGetResp struct {
	Namesapce            *sharedstate.Namespace `json:"namespace"`
	internalResponseMeta `json:"-"`
}

func (n namespacesEndpoint) get(w http.ResponseWriter, r *http.Request) {

	ns := r.Context().Value("name").(string)

	stateResp, err := n.state.Namespaces().Get(&state.NamespacesGetReq{Name: ns})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := NamespaceGetResp{
			Namesapce:            stateResp.Namespace,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type NamespaceListResp struct {
	Namespaces           []*sharedstate.NamespaceStub `json:"namespaces"`
	internalResponseMeta `json:"-"`
}

func (n namespacesEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := n.state.Namespaces().List(&state.NamespacesListReq{})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := NamespaceListResp{
			Namespaces:           stateResp.Namespaces,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

func (n namespacesEndpoint) context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var ns string

		if ns = chi.URLParam(r, "name"); ns == "" {
			httpWriteResponseError(w, errors.New("namespace not found"))
			return
		}

		ctx := context.WithValue(r.Context(), "name", ns) //nolint:staticcheck
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
