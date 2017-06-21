package main

import (
	"github.com/gorilla/mux"
)

func newRouter(controller *Controller) *mux.Router {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", controller.listDatabases).Methods("GET")
	r.HandleFunc("/{db}", controller.listMetrics).Methods("GET")
	r.HandleFunc("/{db}", controller.addSamples).Methods("POST")
	r.HandleFunc("/{db}/_ws", controller.addSamples).Methods("GET")
	r.HandleFunc("/{db}/{metric}", controller.getRawValues).Methods("GET")
	r.HandleFunc("/{db}/{metric}/summary", controller.getSummary).Methods("GET")
	r.HandleFunc("/{db}/{metric}/_heatmap", controller.getHeatMap).Methods("GET")
	r.HandleFunc("/{db}/{metric}/heatmap", controller.getHeatMapSVG).Methods("GET")
	r.HandleFunc("/{db}/{metric}/histo", controller.getHistogram).Methods("GET")

	return r
}
