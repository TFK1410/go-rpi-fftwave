package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//QuadWave ...
type QuadWave struct {
	dataHeight    int
	dataWidth     int
	colBarriers   []float64
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

//InitWave ...
func (m *QuadWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 32
	m.colBarriers = calculateBarriers(m.dataHeight, minVal, maxVal)
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

//Draw ...
func (m *QuadWave) Draw(c *rgbmatrix.Canvas, dmxColor color.RGBA, data, dots []float64, soundEnergyHistory []color.RGBA) {
	bounds := c.Bounds()
	for x, val := range data {
		for y, bar := range m.colBarriers {
			if val > bar {
				if dmxColor.A > 0 {
					m.drawPixels(c, x, y, bounds.Max.X, dmxColor)
				} else {
					m.drawPixels(c, x, y, bounds.Max.X, m.heightColors[y])
				}
			} else {
				m.drawPixels(c, x, y, bounds.Max.X, soundEnergyHistory[m.radiusIndexes[x][y]])
			}
			if dots[x] > bar {
				if y == 0 || m.colBarriers[y-1] > dots[x] {
					m.drawPixels(c, x, y, bounds.Max.X, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
}

func (m *QuadWave) drawPixels(c *rgbmatrix.Canvas, x, y, maxX int, clr color.RGBA) {
	c.Set(m.dataWidth-1-x, y, clr)
	c.Set(m.dataWidth+x, y, clr)
	c.Set(3*m.dataWidth-1-x, m.dataHeight-1-y, clr)
	c.Set(3*m.dataWidth+x, m.dataHeight-1-y, clr)
}
