package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// NoWave defines the values used for the display of the wave that are specific to this pattern type
type NoWave struct {
	dataWidth  int
	dataHeight int
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (nb *NoWave) InitWave(dataWidth int, minVal, maxVal float64) {
	nb.dataWidth = 2 * dataWidth
	nb.dataHeight = 64
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
