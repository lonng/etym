package middleware

import (
	"context"
	"etym/pkg/errutil"
	"etym/pkg/log"
	"net/http"
	"strings"
)

func Logrequest(ctx context.Context, request *http.Request) (context.Context, error) {
	log.Infof("Request => Method=%s, URL=%s, RemoteAddr=%s", request.Method, request.RequestURI, request.RemoteAddr)
	return ctx, nil
}

func LocalFilter(ctx context.Context, request *http.Request) (context.Context, error) {
	parts := strings.Split(request.RemoteAddr, ":")
	if len(parts) != 2 || (parts[0] != "127.0.0.1" && parts[0] != "localhost") {
		return ctx, errutil.ErrPermissionDenied
	}
	return ctx, nil
}
