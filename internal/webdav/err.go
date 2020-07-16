package webdav

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	httpError struct {
		cause  error
		msg    string
		status int
	}

	authRequired struct {
		realm string
	}
)

func (h httpError) Error() string {
	return h.msg
}

func (h httpError) Cause() error {
	return h.cause
}

func (h httpError) PublicError() error {
	return errors.New(h.msg)
}

func (h httpError) Render(w http.ResponseWriter, req *http.Request) {
	http.Error(w, h.PublicError().Error(), h.status)
}

func (a authRequired) Error() string {
	return "request requires authentication"
}

func (a authRequired) Render(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf("Basic realm=%v, charset=utf-8", a.realm))
	w.WriteHeader(http.StatusUnauthorized)
	io.WriteString(w, "please authenticate")
}
