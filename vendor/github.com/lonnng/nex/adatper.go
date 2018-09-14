package nex

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
)

// HandlerAdapter represents a container that contain a handler function
// and convert a it to a http.Handler
type HandlerAdapter interface {
	Invoke(context.Context, http.ResponseWriter, *http.Request) (context.Context, interface{}, error)
}

// genericAdapter represents a common adapter
type genericAdapter struct {
	inContext  bool
	outContext bool
	method     reflect.Value
	numIn      int
	types      []reflect.Type
	cacheArgs  []reflect.Value // cache args
}

// Accept zero parameter adapter
type simplePlainAdapter struct {
	inContext  bool
	outContext bool
	method     reflect.Value
	cacheArgs  []reflect.Value
}

// Accept only one parameter adapter
type simpleUnaryAdapter struct {
	outContext bool
	argType    reflect.Type
	method     reflect.Value
	cacheArgs  []reflect.Value // cache args
}

func makeGenericAdapter(method reflect.Value, inContext, outContext bool) *genericAdapter {
	var noSupportExists = false
	t := method.Type()
	numIn := t.NumIn()

	a := &genericAdapter{
		inContext:  inContext,
		outContext: outContext,
		method:     method,
		numIn:      numIn,
		types:      make([]reflect.Type, numIn),
		cacheArgs:  make([]reflect.Value, numIn),
	}

	for i := 0; i < numIn; i++ {
		in := t.In(i)
		if in != contextType && !isSupportType(in) {
			if noSupportExists {
				panic("function should accept only one customize type")
			}

			if in.Kind() != reflect.Ptr {
				panic("customize type should be a pointer(" + in.PkgPath() + "." + in.Name() + ")")
			}
			noSupportExists = true
		}
		a.types[i] = in
	}

	return a
}

func (a *genericAdapter) Invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (
	outCtx context.Context, payload interface{}, err error) {

	outCtx = ctx
	values := a.cacheArgs
	for i := 0; i < a.numIn; i++ {
		typ := a.types[i]
		v, ok := supportTypes[typ]
		if ok {
			values[i] = v(r)
		} else if typ == contextType {
			values[i] = reflect.ValueOf(ctx)
		} else {
			d := reflect.New(a.types[i].Elem()).Interface()
			err = json.NewDecoder(r.Body).Decode(d)
			if err != nil {
				return
			}
			values[i] = reflect.ValueOf(d)
		}
	}

	ret := a.method.Call(values)

	if a.outContext {
		outCtx = ret[0].Interface().(context.Context)
		payload = ret[1].Interface()
		if e := ret[2].Interface(); e != nil {
			err = e.(error)
		}
	} else {
		payload = ret[0].Interface()
		if e := ret[1].Interface(); e != nil {
			err = e.(error)
		}
	}

	return
}

func (a *simplePlainAdapter) Invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (
	outCtx context.Context, payload interface{}, err error) {
	outCtx = ctx
	if a.inContext {
		a.cacheArgs[0] = reflect.ValueOf(ctx)
	}

	// call it
	ret := a.method.Call(a.cacheArgs)

	if a.outContext {
		outCtx = ret[0].Interface().(context.Context)
		payload = ret[1].Interface()
		if e := ret[2].Interface(); e != nil {
			err = e.(error)
		}
	} else {
		payload = ret[0].Interface()
		if e := ret[1].Interface(); e != nil {
			err = e.(error)
		}
	}

	return
}

func (a *simpleUnaryAdapter) Invoke(ctx context.Context, w http.ResponseWriter, r *http.Request) (
	outCtx context.Context, payload interface{}, err error) {

	outCtx = ctx
	data := reflect.New(a.argType.Elem()).Interface()
	err = json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		return
	}

	a.cacheArgs[0] = reflect.ValueOf(data)
	ret := a.method.Call(a.cacheArgs)

	if a.outContext {
		outCtx = ret[0].Interface().(context.Context)
		payload = ret[1].Interface()
		if e := ret[2].Interface(); e != nil {
			err = e.(error)
		}
	} else {
		payload = ret[0].Interface()
		if e := ret[1].Interface(); e != nil {
			err = e.(error)
		}
	}

	return
}
