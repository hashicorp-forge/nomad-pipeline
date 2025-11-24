package api

import (
	"context"
	"net/http"
)

type Namespace struct {
	ID          string `hcl:"id" json:"id"`
	Description string `hcl:"description,optional" json:"description"`
}

type NamespaceStub struct {
	ID          string
	Description string
}

type Namespaces struct {
	client *Client
}

func (c *Client) Namespaces() *Namespaces {
	return &Namespaces{client: c}
}

type NamespaceCreateReq struct {
	Namespace *Namespace `json:"namespace"`
}

type NamespaceCreateResp struct {
	Namespace *Namespace `json:"namespace"`
}

func (n *Namespaces) Create(ctx context.Context, req *NamespaceCreateReq) (*NamespaceCreateResp, *Response, error) {

	var resp NamespaceCreateResp

	httpReq, err := n.client.NewRequest(http.MethodPost, "/v1/namespaces", req)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := n.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp, httpResp, nil
}

type NamespaceDeleteReq struct {
	Name string `json:"name"`
}

type NamespaceDeleteResp struct{}

func (n *Namespaces) Delete(ctx context.Context, req *NamespaceDeleteReq) (*Response, error) {

	httpReq, err := n.client.NewRequest(http.MethodDelete, "/v1/namespaces/"+req.Name, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := n.client.Do(ctx, httpReq, nil)
	if err != nil {
		return httpResp, err
	}

	return httpResp, nil
}

type NamespaceGetReq struct {
	Name string `json:"name"`
}

type NamespaceGetResp struct {
	Namespace *Namespace `json:"namespace"`
}

func (n *Namespaces) Get(ctx context.Context, req *NamespaceGetReq) (*NamespaceGetResp, *Response, error) {

	var resp NamespaceGetResp

	httpReq, err := n.client.NewRequest(http.MethodGet, "/v1/namespaces/"+req.Name, nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := n.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type NamespaceListReq struct{}

type NamespaceListResp struct {
	Namespaces []*NamespaceStub `json:"namespaces"`
}

func (n *Namespaces) List(ctx context.Context, _ *NamespaceListReq) (*NamespaceListResp, *Response, error) {

	var resp NamespaceListResp

	httpReq, err := n.client.NewRequest(http.MethodGet, "/v1/namespaces", nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := n.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}
