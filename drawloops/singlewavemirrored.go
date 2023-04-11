package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// SingleWaveMirrored defines the values used for the display of the wave that are specific to this pattern type
type SingleWaveMirrored struct {
	dataHeight     int
	dataWidth      int
	minVal, maxVal float64
	paletteIndexes []byte
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *SingleWaveMirrored) InitWave(screenWidth, screenHeight int, minVal, maxVal float64) {
	m.dataWidth = screenWidth
	m.dataHeight = screenHeight / 2
	m.minVal, m.maxVal = minVal, maxVal
	m.paletteIndexes = calculatePaletteIndexes(m.dataHeight)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *SingleWaveMirrored) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64) {
	commonDraw(m, c, dmxData, data, dots)
}

// This function will mirror out a single pixel draw to multiple fields as required
func (m *SingleWaveMirrored) DrawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(x, m.dataHeight*2-1-y, clr)
	c.Set(m.dataWidth-1-x, y, clr)
}

func (m *SingleWaveMirrored) GetDataSize() (int, int) {
	return m.dataWidth, m.dataHeight
}

func (m *SingleWaveMirrored) GetValueRange() (float64, float64) {
	return m.minVal, m.maxVal
}

func (m *SingleWaveMirrored) GetPaletteIndexes() []byte {
	return m.paletteIndexes
}
