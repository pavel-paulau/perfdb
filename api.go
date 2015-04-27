package main

import (
    "io"
	"net/http"
)

type API interface {
    readRequest(r *http.Request) io.ReadCloser
    propagateError(rw http.ResponseWriter, err error, code int)
    validJSON(rw http.ResponseWriter, data interface{})
}
