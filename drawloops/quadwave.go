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
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *QuadWave) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 32
	m.minVal, m.maxVal = minVal, maxVal
	m.paletteIndexes = calculatePaletteIndexes(m.dataHeight)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *QuadWave) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64) {
	for x, val := range data {
		barHeight := getBarHeight(val, m.dataHeight, m.minVal, m.maxVal)
		dotsHeight := getBarHeight(dots[x], m.dataHeight, m.minVal, m.maxVal)
		if dmxData.WhiteDots && barHeight > 0 && barHeight == dotsHeight {
			barHeight--
		}
		var phaseOffset int
		if dmxData.PalettePhaseOffset > 0 && dmxData.PaletteAngle > 0 {
			phaseOffset = int(dmxData.PalettePhaseOffset) + int(float64(dmxData.PaletteAngle)/255.0*float64(m.dataHeight)*float64(x))
		}

		for y := 0; y < barHeight-1; y++ {
			if dmxData.Color.A > 0 {
				// draw constant dmx color
				m.drawPixels(c, x, y, dmxData.Color)
			} else if phaseOffset > 0 && dmxData.ColorPalette > 0 {
				// draw dmx palette color
				m.drawPixels(c, x, y, palette.Palettes[dmxData.ColorPalette][getPaletteOffsetWrap(int(m.paletteIndexes[y])+phaseOffset)])
			} else {
				// draw default palette color
				m.drawPixels(c, x, y, palette.Palettes[0][m.paletteIndexes[y]])
			}
		}

		for y := barHeight; y < m.dataHeight; y++ {
			// blackout the rest
			m.drawPixels(c, x, y, color.RGBA{0, 0, 0, 0})
		}

		if dotsHeight > 0 && dmxData.WhiteDots {
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
