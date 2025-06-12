package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func httpWriteResponse(w http.ResponseWriter, obj interface{}) {

	code := http.StatusInternalServerError

	if respMeta, ok := obj.(internalResponseMeta); ok {
		code = respMeta.StatusCode()
	}

	switch code {
	case http.StatusNoContent:
		w.WriteHeader(code)
		return
	case http.StatusOK:
	default:
		w.WriteHeader(code)
	}

	// If we have a response object, encode it.
	if obj != nil {
		w.Header().Set("Content-Type", "application/json")

		objBytes, err := json.Marshal(obj)
		if err != nil {
			httpWriteResponseError(w, fmt.Errorf("failed to marshal JSON response: %w", err))
			return
		}

		if _, err := w.Write(objBytes); err != nil {
			httpWriteResponseError(w, fmt.Errorf("failed to write JSON response: %w", err))
			return
		}
	}
}

func httpWriteResponseError(w http.ResponseWriter, err error) {
	var (
		code int
		resp []byte
	)

	codedErr, ok := err.(*ResponseError)
	if !ok {
		code = http.StatusInternalServerError
		resp = []byte(err.Error())
	} else {
		code = codedErr.StatusCode()

		objBytes, err := json.Marshal(codedErr)
		if err != nil {
			return
		}
		resp = objBytes
		w.Header().Set("Content-Type", "application/json")
	}

	// Write the status code header.
	w.WriteHeader(code)
	_, _ = w.Write(resp)
}

type internalResponseMeta interface {
	StatusCode() int
}

type internalResponseMetaImpl struct {
	code int
}

func newInternalResponseMeta(c int) internalResponseMetaImpl {
	return internalResponseMetaImpl{
		code: c,
	}
}

func (r internalResponseMetaImpl) StatusCode() int {
	return r.code
}

type ResponseError struct {
	ErrorBody `json:"error"`
}

type ErrorBody struct {
	Msg  string `json:"message"`
	Code int    `json:"code"`
}

func NewResponseError(e error, c int) *ResponseError {
	return &ResponseError{
		ErrorBody: ErrorBody{
			Msg:  e.Error(),
			Code: c,
		},
	}
}

func (e *ResponseError) StatusCode() int { return e.Code }

func (e *ResponseError) Error() string { return e.Msg }

func (e *ResponseError) String() string { return e.Msg }
