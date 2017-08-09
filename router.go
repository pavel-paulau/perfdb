package main

import (
	"github.com/gin-gonic/gin"
)

func newRouter(controller *Controller) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	rg := router.Group("/")

	rg.GET("/", controller.listDatabases)
	rg.GET("/:db", controller.listMetrics)
	rg.GET("/:db/:metric", controller.getRawValues)
	rg.GET("/:db/:metric/summary", controller.getSummary)
	rg.GET("/:db/:metric/heatmap", controller.getHeatMapSVG)

	rg.POST("/:db", controller.addSamples)

	return router
}
