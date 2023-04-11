package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// DualWave defines the values used for the display of the wave that are specific to this pattern type
type DualWave struct {
	dataHeight     int
	dataWidth      int
	minVal, maxVal float64
	paletteIndexes []byte
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *DualWave) InitWave(screenWidth, screenHeight int, minVal, maxVal float64) {
	m.dataWidth = screenWidth / 2
	m.dataHeight = screenHeight
	m.minVal, m.maxVal = minVal, maxVal
	m.paletteIndexes = calculatePaletteIndexes(m.dataHeight)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *DualWave) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64) {
	commonDraw(m, c, dmxData, data, dots)
}

// This function will mirror out a single pixel draw to multiple fields as required
func (m *DualWave) DrawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(2*x, m.dataHeight-1-y, clr)
	c.Set(2*x+1, m.dataHeight-1-y, clr)
}

func (m *DualWave) GetDataSize() (int, int) {
	return m.dataWidth, m.dataHeight
}

func (m *DualWave) GetValueRange() (float64, float64) {
	return m.minVal, m.maxVal
}

func (m *DualWave) GetPaletteIndexes() []byte {
	return m.paletteIndexes
}
