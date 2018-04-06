package main

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/ajstarks/svgo"
)

type margin struct {
	top, right, bottom, left int
}

type size struct {
	width, height int
}

const (
	startFontStyle  = "text-anchor:start;font-size:12px;font-family:Arial,Helvetica"
	middleFontStyle = "text-anchor:middle;font-size:12px;font-family:Arial,Helvetica"
	endFontStyle    = "text-anchor:end;font-size:12px;font-family:Arial,Helvetica"
)

var gridSize = size{6, 10}

func drawCanvas(canvas *svg.SVG, canvasSize size) {
	canvas.Rect(0, 0, canvasSize.width, canvasSize.height,
		"fill:white;stroke:none")
}

func drawXTitle(canvas *svg.SVG, canvasSize, chartInnerSize size, chartMargin margin, timeElapsed time.Duration) {
	var title string

	if timeElapsed.Hours() > 1 {
		title = "Time elapsed, h"
	} else if timeElapsed.Minutes() > 1 {
		title = "Time elapsed, m"
	} else {
		title = "Time elapsed, s"
	}

	canvas.Text(chartMargin.left+chartInnerSize.width/2, canvasSize.height-6,
		title, middleFontStyle)
}

func drawXAxis(canvas *svg.SVG, canvasSize, chartInnerSize size, chartMargin margin, timeElapsed time.Duration) {
	var gridDuration float64

	if timeElapsed.Hours() > 1 {
		gridDuration = timeElapsed.Hours() / float64(gridSize.width)
	} else if timeElapsed.Minutes() > 1 {
		gridDuration = timeElapsed.Minutes() / float64(gridSize.width)
	} else {
		gridDuration = timeElapsed.Seconds() / float64(gridSize.width)
	}

	for i := 0; i <= gridSize.width; i++ {
		tick := fmt.Sprintf("%.1f", float64(i)*gridDuration)
		canvas.Text(chartMargin.left+i*chartInnerSize.width/gridSize.width,
			canvasSize.height-chartMargin.bottom+15,
			tick, middleFontStyle)
	}
}

func drawYTitle(canvas *svg.SVG, chartInnerSize size, chartMargin margin, title string) {
	canvas.Gtransform(fmt.Sprintf("translate(%d,%d) rotate(-90)", 15, chartMargin.top+chartInnerSize.height/2))
	canvas.Text(0, 0, title, middleFontStyle)
	canvas.Gend()
}

func tickFormatter(maxValue float64) string {
	if maxValue > math.Pow(10, 6) {
		return "%.2g"
	} else if maxValue >= math.Pow(10, 3) {
		return "%.0f"
	} else if maxValue < math.Pow(10, -4) {
		return "%.5f"
	}

	for power := -3; power < 1; power++ {
		if maxValue <= math.Pow(10, float64(power)) {
			return fmt.Sprintf("%%.%df", int(math.Abs(float64(power)))+2)
		}
	}

	return "%.1f"
}

func drawYAxis(canvas *svg.SVG, canvasSize, chartInnerSize size, chartMargin margin, hm *heatMap) {
	tickFmt := tickFormatter(hm.MaxValue)
	for i := 0; i <= gridSize.height; i++ {
		tickValue := float64(i) * float64(hm.MaxValue) / float64(gridSize.height)
		tick := fmt.Sprintf(tickFmt, tickValue)
		canvas.Text(chartMargin.left-5,
			canvasSize.height-chartMargin.bottom-i*chartInnerSize.height/gridSize.height,
			tick, endFontStyle)
	}
}

func drawGrid(canvas *svg.SVG, chartInnerSize size, chartMargin margin) {
	// Grid
	const gridStyle = "stroke:black;shape-rendering:crispEdges;stroke-dasharray:2,10"

	for i := 1; i <= gridSize.width-1; i++ {
		canvas.Line(chartMargin.left+i*chartInnerSize.width/gridSize.width,
			chartMargin.top,
			chartMargin.left+i*chartInnerSize.width/gridSize.width,
			chartMargin.top+chartInnerSize.height,
			gridStyle)
	}

	for i := 1; i <= gridSize.height-1; i++ {
		canvas.Line(chartMargin.left,
			chartMargin.top+i*chartInnerSize.height/gridSize.height,
			chartMargin.left+chartInnerSize.width,
			chartMargin.top+i*chartInnerSize.height/gridSize.height,
			gridStyle)
	}

	// Border
	const borderStyle = "fill:none;stroke:black;shape-rendering:crispEdges"
	canvas.Rect(chartMargin.left, chartMargin.top,
		chartInnerSize.width, chartInnerSize.height,
		borderStyle)
}

func drawHeatBar(canvas *svg.SVG, chartInnerSize, chartOuterSize, heatBarInnerSize size, heatBarMargin margin, hm *heatMap) {
	var heatBarColor = []svg.Offcolor{
		{0, "#7F2704", 1.0},
		{25, "#D74701", 1.0},
		{50, "#FC8C3B", 1.0},
		{75, "#FDCFA1", 1.0},
		{100, "#FFFFFF", 1.0},
	}
	canvas.LinearGradient("heatBar", 0, 0, 0, 100, heatBarColor)

	canvas.Rect(chartOuterSize.width+heatBarMargin.left, heatBarMargin.top,
		heatBarInnerSize.width, chartInnerSize.height,
		"fill:url(#heatBar);stroke:black;shape-rendering:crispEdges")

	const heatBarTextMargin = 5

	canvas.Text(chartOuterSize.width+heatBarMargin.left+heatBarInnerSize.width+heatBarTextMargin,
		heatBarMargin.top,
		fmt.Sprintf("%d", hm.maxDensity), startFontStyle)

	canvas.Text(chartOuterSize.width+heatBarMargin.left+heatBarInnerSize.width+heatBarTextMargin,
		heatBarMargin.top+heatBarInnerSize.height,
		"0", startFontStyle)
}

func drawHeatMap(canvas *svg.SVG, canvasSize, chartInnerSize size, chartMargin margin, hm *heatMap) {
	const rectStyle = "fill:%s;stroke:%s"

	for i, row := range hm.Map {
		for j, value := range row {
			idx := math.Pow(float64(value)/float64(hm.maxDensity), 0.15)

			if idx > 0 {
				color := orgColorMap[int(255*idx)]
				canvas.Rect(chartMargin.left+j*chartInnerSize.width/len(row),
					canvasSize.height-chartMargin.bottom-(i+1)*chartInnerSize.height/len(hm.Map),
					chartInnerSize.width/len(row),
					chartInnerSize.height/len(hm.Map),
					fmt.Sprintf(rectStyle, color, color))
			}
		}
	}
}

func generateSVG(output io.Writer, hm *heatMap, title string) {
	// Sizes and margins
	var canvasSize = size{1040, 520}

	var chartMargin = margin{20, 5, 40, 80}
	var heatBarMargin = margin{20, 40, 40, 5}

	var heatBarInnerSize = size{
		18,
		canvasSize.height - heatBarMargin.top - heatBarMargin.bottom,
	}

	var heatBarOuterSize = size{
		heatBarMargin.left + heatBarInnerSize.width + heatBarMargin.right,
		heatBarMargin.top + heatBarInnerSize.height + heatBarMargin.bottom,
	}

	var chartInnerSize = size{
		canvasSize.width - chartMargin.left - chartMargin.right - heatBarOuterSize.width,
		canvasSize.height - chartMargin.top - chartMargin.bottom,
	}

	var chartOuterSize = size{
		chartMargin.left + chartInnerSize.width + chartMargin.right,
		chartMargin.top + chartInnerSize.height + chartMargin.bottom,
	}

	// Drawing
	canvas := svg.New(output)
	canvas.Start(canvasSize.width, canvasSize.height)

	drawCanvas(canvas, canvasSize)

	drawHeatMap(canvas, canvasSize, chartInnerSize, chartMargin, hm)

	timeElapsed := time.Duration(hm.MaxTS-hm.MinTS) * 1e6
	drawXTitle(canvas, canvasSize, chartInnerSize, chartMargin, timeElapsed)
	drawXAxis(canvas, canvasSize, chartInnerSize, chartMargin, timeElapsed)

	drawYAxis(canvas, canvasSize, chartInnerSize, chartMargin, hm)
	drawYTitle(canvas, chartInnerSize, chartMargin, title)

	drawGrid(canvas, chartInnerSize, chartMargin)

	drawHeatBar(canvas, chartInnerSize, chartOuterSize, heatBarInnerSize, heatBarMargin, hm)

	canvas.End()
}
