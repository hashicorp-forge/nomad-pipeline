package dev

import (
	"sync"

	"github.com/oklog/ulid/v2"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type State struct {
	flows     map[flowCompositeKey]*state.Flow
	flowsLock sync.RWMutex

	namespaces     map[string]*state.Namespace
	namesapcesLock sync.RWMutex

	runs     map[runCompositeKey]*state.Run
	runsLock sync.RWMutex

	triggers     map[triggerCompositeKey]*state.Trigger
	triggersLock sync.RWMutex
}

func New() serverstate.State {
	return &State{
		flows:      make(map[flowCompositeKey]*state.Flow),
		namespaces: make(map[string]*state.Namespace),
		runs:       make(map[runCompositeKey]*state.Run),
		triggers:   make(map[triggerCompositeKey]*state.Trigger),
	}
}

type flowCompositeKey struct {
	id        string
	namesapce string
}

type runCompositeKey struct {
	id        ulid.ULID
	namesapce string
}

type triggerCompositeKey struct {
	id        string
	namesapce string
}
