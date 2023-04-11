package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// NoWave defines the values used for the display of the wave that are specific to this pattern type
type NoWave struct {
	dataWidth      int
	dataHeight     int
	minVal, maxVal float64
	paletteIndexes []byte
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (nb *NoWave) InitWave(screenWidth, screenHeight int, minVal, maxVal float64) {
	nb.dataWidth = screenWidth
	nb.dataHeight = screenHeight
}

// Draw creates a new canvas to be later rendered on the matrix
func (nb *NoWave) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64) {
	// Set all matrix pixels to all black
	for x := 0; x < nb.dataWidth; x++ {
		for y := 0; y < nb.dataHeight; y++ {
			c.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
}

// This function will mirror out a single pixel draw to multiple fields as required
func (nb *NoWave) DrawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(2*x, nb.dataHeight-1-y, clr)
	c.Set(2*x+1, nb.dataHeight-1-y, clr)
}

func (nb *NoWave) GetDataSize() (int, int) {
	return nb.dataWidth, nb.dataHeight
}

func (nb *NoWave) GetValueRange() (float64, float64) {
	return nb.minVal, nb.maxVal
}

func (nb *NoWave) GetPaletteIndexes() []byte {
	return nb.paletteIndexes
}
