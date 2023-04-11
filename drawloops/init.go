package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// Wave is used for the implementation of any possible display patterns
type Wave interface {
	InitWave(int, int, float64, float64)
	Draw(*rgbmatrix.Canvas, dmx.DMXData, []float64, []float64)
	DrawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA)
	GetDataSize() (int, int)
	GetValueRange() (float64, float64)
	GetPaletteIndexes() []byte
}

var iterator int
var waves []Wave

// InitWaves creates the wave types array and initializes every one of them
func InitWaves(screenWidth, screenHeight int, minVal, maxVal float64) {
	waves = append(waves, &SingleWave{})
	waves = append(waves, &SingleWaveMirrored{})
	waves = append(waves, &DualWave{})
	waves = append(waves, &MirrorWave{})
	waves = append(waves, &QuadWave{})
	waves = append(waves, &QuadWaveSideways{})
	waves = append(waves, &NoWave{})

	for i := range waves {
		waves[i].InitWave(screenWidth, screenHeight, minVal, maxVal)
	}
}

// GetFirstWave returns the first wave type from the array
func GetFirstWave() Wave {
	return waves[0]
}

func GetWaveNum(i int) Wave {
	iterator = i
	if iterator >= len(waves) {
		iterator = 0
	}
	return waves[iterator]
}

// GetNextWave returns the next wave type from the slice
func GetNextWave() Wave {
	iterator++
	if iterator >= len(waves) {
		iterator = 0
	}
	return waves[iterator]
}
