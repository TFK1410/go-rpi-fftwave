package drawloops

import (
	"image/color"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//Draw ...
func Draw(c *rgbmatrix.Canvas, data []float64) {
	bounds := c.Bounds()
	var xdraw, ydraw int
	for x, val := range data {
		for y, bar := range basicWave.colBarriers {
			if val > bar {
				if y >= len(basicWave.colBarriers)/2 {
					xdraw = x
					ydraw = y - len(basicWave.colBarriers)/2
					c.Set((bounds.Max.X+1)/4*3-1-xdraw, ydraw, color.RGBA{255, 0, 0, 255})
					c.Set((bounds.Max.X+1)/4*3+xdraw, ydraw, color.RGBA{255, 0, 0, 255})
				} else {
					xdraw = x
					ydraw = y
					c.Set((bounds.Max.X+1)/4-1-xdraw, ydraw, color.RGBA{255, 0, 0, 255})
					c.Set((bounds.Max.X+1)/4+xdraw, ydraw, color.RGBA{255, 0, 0, 255})
				}
			}
		}
	}
	c.Render()
}
