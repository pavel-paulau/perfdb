package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bitbucket.org/tebeka/nrsc"
)

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

func validHTML(rw http.ResponseWriter, content string) {
	rw.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(rw, content)
}

func getLineChart(rw http.ResponseWriter, r *http.Request) {
	content, err := readHTML("linechart.html")
	if err != nil {
		logger.Critical(err)
		return
	}
	validHTML(rw, content)
}
