package drawloops

import "fmt"

//LedOutData ...
type LedOutData struct {
	colBarriers        []float64
	dataHeight         int
	whiteDotHeightStep float64
}

var basicWave LedOutData

//InitWaves ...
func InitWaves(minVal, maxVal float64) {
	basicWave.colBarriers = calculateBarriers(64, minVal, maxVal)
	fmt.Println(basicWave.colBarriers)
	basicWave.dataHeight = 64
}
