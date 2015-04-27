package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
