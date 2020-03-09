package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// QuadWaveSideways defines the values used for the display of the wave that are specific to this pattern type
type QuadWaveSideways struct {
	dataHeight    int
	dataWidth     int
	colBarriers   []float64
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *QuadWaveSideways) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 64
	m.colBarriers = calculateBarriers(m.dataHeight, minVal, maxVal)
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *QuadWaveSideways) Draw(c *rgbmatrix.Canvas, dmxColor color.RGBA, data, dots []float64, soundEnergyHistory []color.RGBA) {
	var avg float64
	for x, val := range data {
		if x%2 == 0 {
			avg = val / 2
			continue
		}
		avg += val / 2
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
func (m *QuadWaveSideways) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(m.dataHeight-1-y, m.dataWidth/2-1-x/2, clr)
	c.Set(m.dataHeight+y, m.dataWidth/2-1-x/2, clr)
	c.Set(3*m.dataHeight-1-y, x/2, clr)
	c.Set(3*m.dataHeight+y, x/2, clr)
}
