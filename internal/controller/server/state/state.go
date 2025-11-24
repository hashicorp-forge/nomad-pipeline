package state

import (
	"github.com/oklog/ulid/v2"

	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type State interface {
	Flows() Flows
	Namespaces() Namespaces
	Runs() Runs
	Triggers() Triggers
}

type Flows interface {
	Create(*FlowsCreateReq) (*FlowsCreateResp, *ErrorResp)
	Delete(*FlowsDeleteReq) (*FlowsDeleteResp, *ErrorResp)
	Get(*FlowsGetReq) (*FlowsGetResp, *ErrorResp)
	List(*FlowsListReq) (*FlowsListResp, *ErrorResp)
}

type FlowsCreateReq struct {
	Flow *sharedstate.Flow
}

type FlowsCreateResp struct {
	Flow *sharedstate.Flow
}

type FlowsDeleteReq struct {
	ID        string
	Namespace string
}

type FlowsDeleteResp struct{}

type FlowsGetReq struct {
	ID        string
	Namespace string
}

type FlowsGetResp struct {
	Flow *sharedstate.Flow
}

type FlowsListReq struct {
	Namespace string `json:"namespace"`
}

type FlowsListResp struct {
	Flows []*sharedstate.FlowStub
}

type Namespaces interface {
	Create(*NamespacesCreateReq) (*NamespacesCreateResp, *ErrorResp)
	Delete(*NamespacesDeleteReq) (*NamespacesDeleteResp, *ErrorResp)
	Get(*NamespacesGetReq) (*NamespacesGetResp, *ErrorResp)
	List(*NamespacesListReq) (*NamespacesListResp, *ErrorResp)
}

type NamespacesCreateReq struct {
	Namespace *sharedstate.Namespace
}

type NamespacesCreateResp struct{}

type NamespacesDeleteReq struct {
	Name string
}

type NamespacesDeleteResp struct{}

type NamespacesGetReq struct {
	Name string
}

type NamespacesGetResp struct {
	Namespace *sharedstate.Namespace
}

type NamespacesListReq struct{}

type NamespacesListResp struct {
	Namespaces []*sharedstate.NamespaceStub
}

type Runs interface {
	Create(*RunsCreateReq) (*RunsCreateResp, *ErrorResp)
	Delete(*RunsDeleteReq) (*RunsDeleteResp, *ErrorResp)
	Get(*RunsGetReq) (*RunsGetResp, *ErrorResp)
	List(*RunsListReq) (*RunsListResp, *ErrorResp)
	Update(*RunsUpdateReq) (*RunsUpdateResp, *ErrorResp)
}

type RunsCreateReq struct {
	Run *sharedstate.Run `json:"run"`
}

type RunsCreateResp struct{}

type RunsDeleteReq struct {
	ID        ulid.ULID `json:"id"`
	Namespace string    `json:"namespace"`
}

type RunsDeleteResp struct{}

type RunsGetReq struct {
	ID        ulid.ULID `json:"id"`
	Namespace string    `json:"namespace"`
}

type RunsGetResp struct {
	Run *sharedstate.Run `json:"run"`
}

type RunsListReq struct {
	Namespace string
}

type RunsListResp struct {
	Runs []*sharedstate.RunStub `json:"runs"`
}

type RunsUpdateReq struct {
	Run *sharedstate.Run `json:"run"`
}

type RunsUpdateResp struct{}

type ErrorResp struct {
	ErrorBody `json:"error"`
}

type ErrorBody struct {
	Msg  string `json:"message"`
	Code int    `json:"code"`
	err  error
}

func NewErrorResp(e error, c int) *ErrorResp {
	return &ErrorResp{
		ErrorBody: ErrorBody{
			err:  e,
			Code: c,
			Msg:  e.Error(),
		},
	}
}

func (e *ErrorResp) Error() string { return e.Msg }

func (e *ErrorResp) Err() error { return e.err }

func (e *ErrorResp) StatusCode() int { return e.Code }

func (e *ErrorResp) String() string { return e.Msg }
