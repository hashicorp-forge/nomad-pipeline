package state

type Namespace struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type NamespaceStub struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func (n *Namespace) Stub() *NamespaceStub {
	return &NamespaceStub{
		ID:          n.ID,
		Description: n.Description,
	}
}
