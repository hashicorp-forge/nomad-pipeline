package state

import (
	"errors"
)

type Trigger struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Flow      string `json:"flow"`

	Source *TriggerSource `json:"source"`
}

type TriggerSource struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Config   []byte `json:"config"`
}

func (t *Trigger) Validate() error {

	var errs []error

	if t.ID == "" {
		errs = append(errs, errors.New("trigger ID cannot be empty"))
	}

	if t.Namespace == "" {
		errs = append(errs, errors.New("namespace cannot be empty"))
	}

	if t.Flow == "" {
		errs = append(errs, errors.New("flow cannot be empty"))
	}

	if t.Source == nil {
		errs = append(errs, errors.New("trigger source cannot be empty"))
	}

	return errors.Join(errs...)
}

func (t *Trigger) Stub() *TriggerStub {
	return &TriggerStub{
		ID:        t.ID,
		Namespace: t.Namespace,
		Flow:      t.Flow,
	}
}

type TriggerStub struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Flow      string `json:"flow"`
}
