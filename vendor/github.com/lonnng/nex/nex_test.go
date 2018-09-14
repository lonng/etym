package nex

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testRequest struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}
type testResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var successResponse = &testResponse{Message: "success"}

// acceptable function signature
func withNone() (*testResponse, error)                         { return successResponse, nil }
func withBody(io.ReadCloser) (*testResponse, error)            { return successResponse, nil }
func withReq(*testRequest) (*testResponse, error)              { return successResponse, nil }
func withHeader(http.Header) (*testResponse, error)            { return successResponse, nil }
func withForm(Form) (*testResponse, error)                     { return successResponse, nil }
func withPostForm(PostForm) (*testResponse, error)             { return successResponse, nil }
func withFormPtr(*Form) (*testResponse, error)                 { return successResponse, nil }
func withPostFormPtr(*PostForm) (*testResponse, error)         { return successResponse, nil }
func withMultipartForm(*multipart.Form) (*testResponse, error) { return successResponse, nil }
func withUrl(*url.URL) (*testResponse, error)                  { return successResponse, nil }
func withRawRequest(*http.Request) (*testResponse, error)      { return successResponse, nil }

func withInContext(context.Context) (*testResponse, error) { return successResponse, nil }

func withInContextAndPayload(context.Context, *testRequest) (*testResponse, error) {
	return successResponse, nil
}

func withOutContext() (context.Context, *testResponse, error) {
	return context.Background(), successResponse, nil
}

func withMulti(*testRequest, Form, PostForm, http.Header, *url.URL) (*testResponse, error) {
	return nil, nil
}
func withAll(io.ReadCloser, *testRequest, Form, PostForm, http.Header, *multipart.Form, *url.URL) (*testResponse, error) {
	return nil, nil
}

func TestHandler(t *testing.T) {
	Handler(withNone)
	Handler(withBody)
	Handler(withReq)
	Handler(withHeader)
	Handler(withForm)
	Handler(withPostForm)
	Handler(withFormPtr)
	Handler(withPostFormPtr)
	Handler(withMultipartForm)
	Handler(withUrl)
	Handler(withRawRequest)
	Handler(withMulti)
	Handler(withAll)
	Handler(withInContext)
	Handler(withOutContext)
	Handler(withInContextAndPayload)
}

func TestBefore(t *testing.T) {
	logic := func(ctx context.Context) (*testResponse, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}
		if ctx.Value("key2").(string) != "value2" {
			t.Fail()
		}
		return &testResponse{}, nil
	}
	before1 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		return context.WithValue(ctx, "key", "value"), nil
	}

	before2 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}

		return context.WithValue(ctx, "key2", "value2"), nil
	}

	handler := Handler(logic).Before(before1, before2)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(recorder, request)
}

func TestAfter(t *testing.T) {
	logic := func(ctx context.Context) (context.Context, *testResponse, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}
		if ctx.Value("key2").(string) != "value2" {
			t.Fail()
		}

		return context.WithValue(ctx, "logic", "logic-value"), &testResponse{}, nil
	}

	before1 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		return context.WithValue(ctx, "key", "value"), nil
	}

	before2 := func(ctx context.Context, request *http.Request) (context.Context, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}
		return context.WithValue(ctx, "key2", "value2"), nil
	}

	after1 := func(ctx context.Context, w http.ResponseWriter) (context.Context, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}
		if ctx.Value("key2").(string) != "value2" {
			t.Fail()
		}
		if ctx.Value("logic").(string) != "logic-value" {
			t.Fail()
		}

		return context.WithValue(ctx, "after1", "after1-value"), nil
	}

	after2 := func(ctx context.Context, w http.ResponseWriter) (context.Context, error) {
		if ctx.Value("key").(string) != "value" {
			t.Fail()
		}
		if ctx.Value("key2").(string) != "value2" {
			t.Fail()
		}
		if ctx.Value("logic").(string) != "logic-value" {
			t.Fail()
		}
		if ctx.Value("after1").(string) != "after1-value" {
			t.Fail()
		}

		return context.WithValue(ctx, "key", "value"), nil
	}

	handler := Handler(logic).Before(before1, before2).After(after1, after2)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(recorder, request)
}

func BenchmarkSimplePlainAdapter_Invoke(b *testing.B) {
	handler := Handler(withNone)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkSimpleUnaryAdapter_Invoke(b *testing.B) {
	handler := Handler(withReq)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}

func BenchmarkGenericAdapter_Invoke(b *testing.B) {
	handler := Handler(withMulti)
	request, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		b.Fatal(err)
	}
	payload := []byte(`{"for":"hello", "bar":10000}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
	}
}
