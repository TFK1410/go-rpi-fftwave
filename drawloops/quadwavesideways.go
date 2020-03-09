package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//QuadWaveSideways ...
type QuadWaveSideways struct {
	dataHeight    int
	dataWidth     int
	colBarriers   []float64
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

//InitWave ...
func (m *QuadWaveSideways) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 64
	m.colBarriers = calculateBarriers(m.dataHeight, minVal, maxVal)
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

//Draw ...
func (m *QuadWaveSideways) Draw(c *rgbmatrix.Canvas, dmxColor color.RGBA, data, dots []float64, soundEnergyHistory []color.RGBA) {
	var avg float64
	for x, val := range data {
		if x%2 == 0 {
			avg = val / 2
			continue
		}
		avg += val / 2
		for y, bar := range m.colBarriers {
			if avg > bar {
				if dmxColor.A > 0 {
					m.drawPixels(c, x, y, dmxColor)
				} else {
					m.drawPixels(c, x, y, m.heightColors[y])
				}
			} else {
				m.drawPixels(c, x, y, soundEnergyHistory[m.radiusIndexes[x][y]])
			}
			if dots[x] > bar {
				if y == 0 || m.colBarriers[y-1] > dots[x] {
					m.drawPixels(c, x, y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
}

func (m *QuadWaveSideways) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(m.dataHeight-1-y, m.dataWidth/2-1-x/2, clr)
	c.Set(m.dataHeight+y, m.dataWidth/2-1-x/2, clr)
	c.Set(3*m.dataHeight-1-y, x/2, clr)
	c.Set(3*m.dataHeight+y, x/2, clr)
}
