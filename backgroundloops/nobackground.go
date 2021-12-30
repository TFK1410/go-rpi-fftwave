package backgroundloops

import (
	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// MirrorWave defines the values used for the display of the wave that are specific to this pattern type
type NoBackground struct {
}

// InitWave does the initial calculation of the reused variables in the draw loop
func (b *NoBackground) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
}

// Draw adds the background details to the canvas on the matrix
func (m *NoBackground) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
}
