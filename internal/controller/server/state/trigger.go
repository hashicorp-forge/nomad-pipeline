package state

import sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"

type Triggers interface {
	Create(*TriggersCreateReq) (*TriggersCreateResp, *ErrorResp)
	Delete(*TriggersDeleteReq) (*TriggersDeleteResp, *ErrorResp)
	Get(*TriggersGetReq) (*TriggersGetResp, *ErrorResp)
	List(*TriggersListReq) (*TriggersListResp, *ErrorResp)
}

type TriggersCreateReq struct {
	Trigger *sharedstate.Trigger
}

type TriggersCreateResp struct{}

type TriggersDeleteReq struct {
	ID        string
	Namespace string
}

type TriggersDeleteResp struct{}

type TriggersGetReq struct {
	ID        string
	Namespace string
}

type TriggersGetResp struct {
	Trigger *sharedstate.Trigger
}

type TriggersListReq struct {
	Namespace string
}

type TriggersListResp struct {
	Triggers []*sharedstate.TriggerStub
}
