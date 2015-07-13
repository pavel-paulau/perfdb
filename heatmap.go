package main

type heatMap struct {
	MinTS      int64   `json:"minTimestamp"`
	MaxTS      int64   `json:"maxTimestamp"`
	MaxValue   float64 `json:"maxValue"`
	Map        [][]int `json:"map"`
	maxDensity int     // Private field
}

const (
	heatMapHeight = 120
	heatMapWidth  = 240
)

func newHeatMap() *heatMap {
	hm := heatMap{}
	hm.Map = [][]int{}
	for y := 0; y < heatMapHeight; y++ {
		hm.Map = append(hm.Map, []int{})
		for x := 0; x < heatMapWidth; x++ {
			hm.Map[y] = append(hm.Map[y], 0)
		}
	}
	return &hm
}
