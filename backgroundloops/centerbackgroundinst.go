package backgroundloops

import (
	"image/color"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// CenterBackgroundInst defines the values used for the display of the wave that are specific to this pattern type
type CenterBackgroundInst struct {
	dataWidth, dataHeight int
	min, max              float64
	hueRotation           time.Duration
	radiusIndexes         [][]int
	centerX, centerY      float64
	height                int
}

// InitBackgroundLoop does the initial calculation of the reused variables in the draw loop
func (cbi *CenterBackgroundInst) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
	cbi.dataWidth = displayWidth
	cbi.dataHeight = displayHeight
	cbi.min = minVal
	cbi.max = maxVal
	if cbi.hueRotation == 0 {
		cbi.hueRotation = 10 * time.Second
	}
	if cbi.centerX == 0 && cbi.centerY == 0 {
		cbi.centerX = float64(displayWidth)/2 - 0.5
		cbi.centerY = float64(displayHeight)/2 - 0.5
	}
	cbi.radiusIndexes = calculateDistance(displayWidth, displayHeight, cbi.centerX, cbi.centerY)

	// get max radiusIndex
	mx := 0
	for x := 0; x < displayWidth; x++ {
		for y := 0; y < displayHeight; y++ {
			if cbi.radiusIndexes[x][y] > mx {
				mx = cbi.radiusIndexes[x][y]
			}
		}
	}
	cbi.height = mx - 1
}

// Draw adds the background details to the canvas on the matrix
func (cbi *CenterBackgroundInst) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
	var H, S, V, soundEnergy float64
	var energyHeight int
	var clr color.RGBA
	H = float64(time.Now().UnixMilli()%cbi.hueRotation.Milliseconds()) / float64(cbi.hueRotation.Milliseconds())
	S = 1
	V = 0.3
	clr = hsv2RGB(H, S, V)

	soundEnergy = 0
	for i := 0; i < 5; i++ {
		soundEnergy += maxTriBand(soundHistory[i]) / 5
	}
	energyHeight = int((soundEnergy - float64(cbi.min)) / float64((cbi.max - cbi.min)) * float64(cbi.height))

	for y := 0; y < cbi.dataHeight; y++ {
		for x := 0; x < cbi.dataWidth; x++ {
			_, _, _, a := c.At(x, y).RGBA()
			if a == 0 && cbi.radiusIndexes[x][y] < energyHeight {
				c.Set(x, y, clr)
			}
		}
	}
}
