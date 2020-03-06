package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

type wave interface {
	InitWave(int, float64, float64)
	Draw(*rgbmatrix.Canvas, []float64, []float64, []color.RGBA)
}

//BasicWave ...
var BasicWave MirrorWave

//InitWaves ...
func InitWaves(dataWidth int, minVal, maxVal float64) {
	BasicWave.InitWave(dataWidth, minVal, maxVal)
}
