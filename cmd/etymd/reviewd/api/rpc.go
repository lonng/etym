package api

import (
	"etym/pkg/constant"
	"etym/pkg/log"
	"etym/pkg/token"
	"io"
	"net/http"
	"strings"
)

var RemoteUrl string

func Proxy(w http.ResponseWriter, r *http.Request) {
	t := strings.TrimSpace(r.Header.Get(constant.Authorization))
	if t == "" {
		//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	meta, err := token.Parse(t)
	if err != nil || meta.Account != constant.Superadmin {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var resp *http.Response
	var remote = RemoteUrl + r.RequestURI
	log.Infof("Proxy => Method=%v, RemoteAddr=%s, URL=%s", r.Method, r.RemoteAddr, remote)
	if r.Method == http.MethodGet {
		resp, err = http.Get(remote)
	} else if r.Method == http.MethodPost {
		resp, err = http.Post(remote, "application/json", r.Body)
	}
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Error(err)
	}
}
