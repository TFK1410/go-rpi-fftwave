package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	"github.com/TFK1410/go-rpi-fftwave/palette"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// QuadWave defines the values used for the display of the wave that are specific to this pattern type
type QuadWave struct {
	dataHeight     int
	dataWidth      int
	minVal, maxVal float64
	paletteIndexes []byte
	radiusIndexes  [][]int
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *QuadWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 32
	m.minVal, m.maxVal = minVal, maxVal
	m.paletteIndexes = calculatePaletteIndexes(m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *QuadWave) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64, soundEnergyHistory []color.RGBA) {
	for x, val := range data {
		barHeight := getBarHeight(val, m.dataHeight, m.minVal, m.maxVal)
		dotsHeight := getBarHeight(dots[x], m.dataHeight, m.minVal, m.maxVal)
		if (!dmxData.DMXOn || dmxData.WhiteDots) && barHeight > 0 && barHeight == dotsHeight {
			barHeight--
		}
		var phaseOffset byte
		if dmxData.DMXOn {
			phaseOffset = dmxData.PalettePhaseOffset + byte(float64(dmxData.PaletteAngle)/255.0*float64(m.dataHeight)*float64(x))
		}

		for y := 0; y < barHeight-1; y++ {
			if dmxData.DMXOn {
				if dmxData.Color.A > 0 {
					// draw constant dmx color
					m.drawPixels(c, x, y, dmxData.Color)
				} else {
					// draw dmx palette color
					m.drawPixels(c, x, y, palette.Palettes[dmxData.ColorPalette][m.paletteIndexes[y]+phaseOffset])
				}
			} else {
				// draw default palette color
				m.drawPixels(c, x, y, palette.Palettes[0][m.paletteIndexes[y]])
			}
		}

		for y := barHeight; y < m.dataHeight; y++ {
			// sound energy color draw
			m.drawPixels(c, x, y, soundEnergyHistory[m.radiusIndexes[x][y]])
		}

		if dotsHeight > 0 && (!dmxData.DMXOn || dmxData.WhiteDots) {
			// white dot draw
			m.drawPixels(c, x, dotsHeight-1, color.RGBA{255, 255, 255, 255})
		}
	}
}

// This function will mirror out a single pixel draw to multiple fields as required
func (m *QuadWave) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(m.dataWidth-1-x, m.dataHeight-1-y, clr)
	c.Set(m.dataWidth+x, m.dataHeight-1-y, clr)
	c.Set(m.dataWidth-1-x, m.dataHeight+y, clr)
	c.Set(m.dataWidth+x, m.dataHeight+y, clr)
}
