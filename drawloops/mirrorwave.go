package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//MirrorWave ...
type MirrorWave struct {
	colBarriers   []float64
	dataHeight    int
	heightColors  []color.RGBA
	radiusIndexes [][]int
}

//InitWave ...
func (m *MirrorWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.colBarriers = calculateBarriers(64, minVal, maxVal)
	m.dataHeight = 64
	m.heightColors = colorGradient(color.RGBA{0, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, m.dataHeight)
	m.radiusIndexes = calculateDistance(dataWidth, m.dataHeight, -0.5, float64(m.dataHeight+1)/2)
}

//Draw ...
func (m *MirrorWave) Draw(c *rgbmatrix.Canvas, data, dots []float64, soundEnergyHistory []color.RGBA) {
	bounds := c.Bounds()
	for x, val := range data {
		for y, bar := range m.colBarriers {
			if val > bar {
				m.drawPixels(c, x, y, bounds.Max.X, m.heightColors[y])
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
	c.Render()
}

func (m *MirrorWave) drawPixels(c *rgbmatrix.Canvas, x, y, maxX int, clr color.RGBA) {
	if y >= len(m.colBarriers)/2 {
		c.Set((maxX+1)/4*3-1-x, y-len(m.colBarriers)/2, clr)
		c.Set((maxX+1)/4*3+x, y-len(m.colBarriers)/2, clr)
	} else {
		c.Set((maxX+1)/4-1-x, y, clr)
		c.Set((maxX+1)/4+x, y, clr)
	}
}
