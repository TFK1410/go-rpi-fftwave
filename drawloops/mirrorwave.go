package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//MirrorWave ...
type MirrorWave struct {
	dataHeight    int
	dataWidth     int
	colBarriers   []float64
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

//InitWave ...
func (m *MirrorWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 64
	m.colBarriers = calculateBarriers(m.dataHeight, minVal, maxVal)
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight+1)/2)
}

//Draw ...
func (m *MirrorWave) Draw(c *rgbmatrix.Canvas, dmxColor color.RGBA, data, dots []float64, soundEnergyHistory []color.RGBA) {
	for x, val := range data {
		for y, bar := range m.colBarriers {
			if val > bar {
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

func (m *MirrorWave) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	if y < m.dataHeight/2 {
		c.Set(m.dataWidth-1-x, y, clr)
		c.Set(m.dataWidth+x, y, clr)
	} else {
		c.Set(3*m.dataWidth-1-x, y-m.dataHeight/2, clr)
		c.Set(3*m.dataWidth+x, y-m.dataHeight/2, clr)
	}
}
