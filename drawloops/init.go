package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//Wave ...
type Wave interface {
	InitWave(int, float64, float64)
	Draw(*rgbmatrix.Canvas, color.RGBA, []float64, []float64, []color.RGBA)
}

var iterator int
var waves []Wave

//InitWaves ...
func InitWaves(dataWidth int, minVal, maxVal float64) {
	waves = append(waves, &MirrorWave{})
	waves = append(waves, &QuadWave{})
	waves = append(waves, &DualWave{})

	for i := range waves {
		waves[i].InitWave(dataWidth, minVal, maxVal)
	}
}

//GetFirstWave ...
func GetFirstWave() Wave {
	return waves[0]
}

//GetNextWave ...
func GetNextWave() Wave {
	iterator++
	if iterator >= len(waves) {
		iterator = 0
	}
	return waves[iterator]
}
