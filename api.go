package main

type API interface {
	readJSON() (interface{}, error)
	writeError(err error, code int)
	writeJSON(data interface{})
	open() error
}
