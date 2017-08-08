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
		b, _ = json.Marshal(map[string]string{
			"error": err.Error(),
		})
	}
	s = string(b)
	return
}

type restHandler struct {
	rw http.ResponseWriter
	r  *http.Request
}

func (rest *restHandler) open() error {
	return nil
}

func (rest *restHandler) readJSON() (interface{}, error) {
	var data interface{}

	decoder := json.NewDecoder(rest.r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (rest *restHandler) writeError(err error, code int) {
	logger.Critical(err)
	rest.rw.Header().Set("Content-Type", "application/json")
	switch code {
	case 400:
		rest.rw.WriteHeader(http.StatusBadRequest)
	case 404:
		rest.rw.WriteHeader(http.StatusNotFound)
	case 500:
		rest.rw.WriteHeader(http.StatusInternalServerError)
	}
	resp := map[string]string{"error": err.Error()}
	fmt.Fprint(rest.rw, httpResponse{resp})
}

func (rest *restHandler) writeJSON(data interface{}) {
	rest.rw.Header().Set("Content-Type", "application/json")
	fmt.Fprint(rest.rw, httpResponse{data})
}
