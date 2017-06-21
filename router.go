package main

import (
	"github.com/gorilla/mux"
)

func newRouter(controller *Controller) *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", controller.listDatabases).Methods("GET")
	r.HandleFunc("/{db}", controller.listSources).Methods("GET")
	r.HandleFunc("/{db}/{source}", controller.listMetrics).Methods("GET")
	r.HandleFunc("/{db}/{source}", controller.addSamples).Methods("POST")
	r.HandleFunc("/{db}/{source}/_ws", controller.addSamples).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}", controller.getRawValues).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/summary", controller.getSummary).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/_heatmap", controller.getHeatMap).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/heatmap", controller.getHeatMapSVG).Methods("GET")
	r.HandleFunc("/{db}/{source}/{metric}/histo", controller.getHistogram).Methods("GET")

	return r
}
