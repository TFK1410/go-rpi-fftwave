package drawloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	"github.com/TFK1410/go-rpi-fftwave/palette"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// QuadWaveSideways defines the values used for the display of the wave that are specific to this pattern type
type QuadWaveSideways struct {
	dataHeight     int
	dataWidth      int
	minVal, maxVal float64
	paletteIndexes []byte
	radiusIndexes  [][]int
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (m *QuadWaveSideways) InitWave(dataWidth int, minVal, maxVal float64) {
	m.dataWidth = dataWidth
	m.dataHeight = 64
	m.minVal, m.maxVal = minVal, maxVal
	m.paletteIndexes = calculatePaletteIndexes(m.dataHeight)
	m.radiusIndexes = calculateDistance(m.dataWidth, m.dataHeight, -0.5, float64(m.dataHeight)+0.5)
}

// Draw creates a new canvas to be later rendered on the matrix
func (m *QuadWaveSideways) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64, soundEnergyHistory []color.RGBA) {
	var avg, avgDots float64
	for x, val := range data {
		if x%2 == 0 {
			avg = val / 2
			avgDots = dots[x] / 2
			continue
		}
		avg += val / 2
		avgDots += dots[x] / 2

		barHeight := getBarHeight(avg, m.dataHeight, m.minVal, m.maxVal)
		dotsHeight := getBarHeight(avgDots, m.dataHeight, m.minVal, m.maxVal)
		if (!dmxData.DMXOn || dmxData.WhiteDots) && barHeight > 0 && barHeight == dotsHeight {
			barHeight--
		}
		var phaseOffset int
		if dmxData.DMXOn {
			phaseOffset = int(dmxData.PalettePhaseOffset) + int(float64(dmxData.PaletteAngle)/255.0*float64(m.dataHeight)*float64(x))
		}

		for y := 0; y < barHeight-1; y++ {
			if dmxData.DMXOn {
				if dmxData.Color.A > 0 {
					// draw constant dmx color
					m.drawPixels(c, x, y, dmxData.Color)
				} else {
					// draw dmx palette color
					m.drawPixels(c, x, y, palette.Palettes[dmxData.ColorPalette][getPaletteOffsetWrap(int(m.paletteIndexes[y])+phaseOffset)])
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
func (m *QuadWaveSideways) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	c.Set(y, m.dataWidth/2-1-x/2, clr)
	c.Set(m.dataHeight*2-1-y, m.dataWidth/2-1-x/2, clr)
	c.Set(y, (m.dataWidth+x)/2, clr)
	c.Set(m.dataHeight*2-1-y, (m.dataWidth+x)/2, clr)
}
