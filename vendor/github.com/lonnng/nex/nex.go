package nex

import (
	"context"
	"encoding/json"
	"net/http"
)

type (
	// ErrorEncoder encode error to response body
	ErrorEncoder func(error) interface{}

	// ResponseEncoder encode payload to response body
	ResponseEncoder func(payload interface{}) interface{}

	// StatusCodeEncoder encode error to response status code
	StatusCodeEncoder func(error) int

	// DefaultErrorMessage wrap error to json
	DefaultErrorMessage struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}

	// Nex represents a handler that contains a bundle of hooks
	Nex struct {
		before  []BeforeFunc
		after   []AfterFunc
		adapter HandlerAdapter
	}

	// NexGroup represents a handler group that contains same hooks
	NexGroup struct {
		before []BeforeFunc
		after  []AfterFunc
	}
)

var (
	errorEncoder      ErrorEncoder
	responseEncoder   ResponseEncoder
	statusCodeEncoder StatusCodeEncoder
)

func fail(w http.ResponseWriter, err error) {
	w.WriteHeader(statusCodeEncoder(err))
	json.NewEncoder(w).Encode(errorEncoder(err))
}

func succ(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(responseEncoder(data))
}

func (n *Nex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var (
		ctx  context.Context = context.Background()
		err  error
		resp interface{}
	)
	// global before middleware
	for _, b := range globalBefore {
		ctx, err = b(ctx, r)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// before middleware
	for _, b := range n.before {
		ctx, err = b(ctx, r)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// adapter handler
	ctx, resp, err = n.adapter.Invoke(ctx, w, r)
	if err != nil {
		fail(w, err)
		return
	}

	// after middleware
	for _, a := range n.after {
		ctx, err = a(ctx, w)
		if err != nil {
			fail(w, err)
			return
		}
	}

	// global after middleware
	for _, a := range globalAfter {
		ctx, err = a(ctx, w)
		if err != nil {
			fail(w, err)
			return
		}
	}
	if err != nil {
		fail(w, err)
	} else {
		succ(w, resp)
	}
}

func (n *Nex) Before(before ...BeforeFunc) *Nex {
	for _, b := range before {
		if b != nil {
			n.before = append(n.before, b)
		}
	}
	return n
}

func (n *Nex) After(after ...AfterFunc) *Nex {
	for _, a := range after {
		if a != nil {
			n.after = append(n.after, a)
		}
	}
	return n
}

func init() {
	errorEncoder = func(err error) interface{} {
		return &DefaultErrorMessage{
			Code:  -1,
			Error: err.Error(),
		}
	}

	responseEncoder = func(payload interface{}) interface{} {
		return payload
	}

	statusCodeEncoder = func(error) int {
		return http.StatusBadRequest
	}
}
