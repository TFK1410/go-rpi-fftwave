package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// DualWave defines the values used for the display of the wave that are specific to this pattern type
type DualWave struct {
	dataHeight    int
	dataWidth     int
	colBarriers   []float64
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *DualWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 64
	m.colBarriers = calculateBarriers(m.dataHeight, minVal, maxVal)
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *DualWave) Draw(c *rgbmatrix.Canvas, dmxColor color.RGBA, data, dots []float64, soundEnergyHistory []color.RGBA) {
	for x, val := range data {
		for y, bar := range m.colBarriers {
			if val > bar {
				if dmxColor.A > 0 {
					// DMX Color draw
					m.drawPixels(c, x, y, dmxColor)
				} else {
					// usual FFT bar draw
					m.drawPixels(c, x, y, m.heightColors[y])
				}
			} else {
				// sound energy color draw
				m.drawPixels(c, x, y, soundEnergyHistory[m.radiusIndexes[x][y]])
			}
			if dots[x] > bar {
				if y == 0 || m.colBarriers[y-1] > dots[x] {
					// white dot draw
					m.drawPixels(c, x, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
}

// This function will mirror out a single pixel draw to multiple fields as required
func (m *DualWave) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	x1 := 2 * x
	x2 := 2*x + 1
	if y < m.dataHeight/2 {
		c.Set(x1, y, clr)
		c.Set(x2, y, clr)
	} else {
		c.Set(2*m.dataWidth+x1, y-m.dataHeight/2, clr)
		c.Set(2*m.dataWidth+x2, y-m.dataHeight/2, clr)
	}
}
