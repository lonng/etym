package nex

import "reflect"

func Handler(f interface{}) *Nex {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("invalid parameter")
	}

	numOut := t.NumOut()

	if numOut != 2 && numOut != 3 {
		panic("unsupported function type, function return values should contain response data & error")
	}

	if numOut == 3 {
		if t.Out(0) != contextType {
			panic("unsupported function type")
		}
	}

	var (
		adapter    HandlerAdapter
		numIn      = t.NumIn()
		outContext = numOut == 3
		inContext  = false
	)

	if numIn > 0 {
		for i := 0; i < numIn; i++ {
			if t.In(i) == contextType {
				inContext = true
			}
		}
	}

	if numIn == 0 {
		adapter = &simplePlainAdapter{
			inContext:  false,
			outContext: outContext,
			method:     reflect.ValueOf(f),
			cacheArgs:  []reflect.Value{},
		}
	} else if numIn == 1 && inContext {
		adapter = &simplePlainAdapter{
			inContext:  true,
			outContext: outContext,
			method:     reflect.ValueOf(f),
			cacheArgs:  make([]reflect.Value, 1),
		}
	} else if numIn == 1 && !isSupportType(t.In(0)) && t.In(0).Kind() == reflect.Ptr {
		adapter = &simpleUnaryAdapter{
			outContext: outContext,
			argType:    t.In(0),
			method:     reflect.ValueOf(f),
			cacheArgs:  make([]reflect.Value, 1),
		}
	} else {
		adapter = makeGenericAdapter(reflect.ValueOf(f), inContext, outContext)
	}

	return &Nex{adapter: adapter}
}

func SetErrorEncoder(c ErrorEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	errorEncoder = c
}

func SetResponseEncoder(c ResponseEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	responseEncoder = c
}

func SetStatusCodeEncoder(c StatusCodeEncoder) {
	if c == nil {
		panic("nil pointer to error encoder")
	}
	statusCodeEncoder = c
}

func SetMultipartFormMaxMemory(m int64) {
	maxMemory = m
}
