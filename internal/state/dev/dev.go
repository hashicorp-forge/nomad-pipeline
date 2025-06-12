package dev

import (
	"sync"
	
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type State struct {
	flows     map[string]*state.Flow
	flowsLock sync.RWMutex

	runs     map[ulid.ULID]*state.Run
	runsLock sync.RWMutex
}

func New() state.State {
	return &State{
		flows: make(map[string]*state.Flow),
		runs:  make(map[ulid.ULID]*state.Run),
	}
}
