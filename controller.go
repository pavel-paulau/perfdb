package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	storage *perfDB
}

func newController(storage *perfDB) *Controller {
	return &Controller{storage}
}

func (c *Controller) listDatabases(context *gin.Context) {
	databases, err := c.storage.listDatabases()
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	context.JSON(http.StatusOK, databases)
}

func (c *Controller) checkDbExists(dbname string) error {
	status, err := c.storage.isExist(dbname)
	if err != nil {
		return err
	}
	if !status {
		return errors.New("not found")
	}
	return nil
}

func (c *Controller) listMetrics(context *gin.Context) {
	dbname := context.Param("db")

	if err := c.checkDbExists(dbname); err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}

	metrics, err := c.storage.listMetrics(dbname)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, metrics)
}

func (c *Controller) getRawValues(context *gin.Context) {
	dbname := context.Param("db")
	metric := context.Param("metric")

	if err := c.checkDbExists(dbname); err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}

	values, err := c.storage.getRawValues(dbname, metric)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, values)
}

func (c *Controller) getSummary(context *gin.Context) {
	dbname := context.Param("db")
	metric := context.Param("metric")

	if err := c.checkDbExists(dbname); err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}

	values, err := c.storage.getSummary(dbname, metric)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, values)
}

func (c *Controller) addSamples(context *gin.Context) {
	var tsNano int64
	if timestamp := context.Query("ts"); timestamp != "" {
		tsNano = parseTimestamp(timestamp)
	} else {
		tsNano = time.Now().UnixNano()
	}

	dbname := context.Param("db")

	var samples map[string]interface{}
	if err := context.BindJSON(&samples); err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}

	for metric, value := range samples {
		sample := Sample{tsNano, value.(float64)}
		if err := c.storage.addSample(dbname, metric, sample); err != nil {
			context.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	context.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (c *Controller) getHeatMapSVG(context *gin.Context) {
	dbname := context.Param("db")
	metric := context.Param("metric")

	if err := c.checkDbExists(dbname); err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}

	hm, err := c.storage.getHeatMap(dbname, metric)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var title string
	if label := context.Query("label"); label != "" {
		title = label
	} else {
		title = metric
	}

	context.Writer.Header().Set("Content-Type", "image/svg+xml")
	generateSVG(context.Writer, hm, title)
}
