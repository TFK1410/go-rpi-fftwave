package drawloops

import (
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

type wave interface {
	InitWave(int, float64, float64, float64, float64)
	Draw(*rgbmatrix.Canvas, []float64, []float64)
}

//BasicWave ...
var BasicWave MirrorWave

//InitWaves ...
func InitWaves(dataWidth int, minVal, maxVal, minEnergy, maxEnergy float64) {
	BasicWave.InitWave(dataWidth, minVal, maxVal, minEnergy, maxEnergy)
}
