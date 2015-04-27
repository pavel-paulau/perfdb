package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bitbucket.org/tebeka/nrsc"
)

type httpResponse struct {
	Raw interface{}
}

func (r httpResponse) String() (s string) {
	b, err := json.Marshal(r.Raw)
	if err != nil {
		b, err = json.Marshal(map[string]string{
			"error": fmt.Sprint("", err),
		})
	}
	s = string(b)
	return
}

func readHTML(path string) (string, error) {
	var html nrsc.Resource
	if html = nrsc.Get(path); html == nil {
		return "", errors.New("cannot read HTML")
	}
	var htmlReader io.Reader
	var err error
	if htmlReader, err = html.Open(); err != nil {
		return "", err
	}
	var content []byte
	if content, err = ioutil.ReadAll(htmlReader); err != nil {
		return "", err
	}
	return string(content), nil
}

func propagateError(rw http.ResponseWriter, err error, code int) {
	logger.Critical(err)
	rw.Header().Set("Content-Type", "application/json")
	switch code {
	case 400:
		rw.WriteHeader(http.StatusBadRequest)
	case 404:
		rw.WriteHeader(http.StatusNotFound)
	case 500:
		rw.WriteHeader(http.StatusInternalServerError)
	}
	resp := map[string]string{"error": err.Error()}
	fmt.Fprint(rw, httpResponse{resp})
}

func validJSON(rw http.ResponseWriter, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rw, httpResponse{data})
}

func validHTML(rw http.ResponseWriter, content string) {
	rw.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(rw, content)
}
