package backgroundloops

import (
	"image/color"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// DesaturateBackground defines the values used for the display of the wave that are specific to this pattern type
type DesaturateBackground struct {
	dataWidth, dataHeight int
	min, max              float64
}

// InitBackgroundLoop does the initial calculation of the reused variables in the draw loop
func (db *DesaturateBackground) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
	db.dataWidth = displayWidth
	db.dataHeight = displayHeight
	db.min = minVal
	db.max = maxVal
}

// Draw adds the background details to the canvas on the matrix
func (db *DesaturateBackground) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
	var soundEnergy, energyDesat float64

	soundEnergy = 0
	for i := 0; i < 5; i++ {
		soundEnergy += maxTriBand(soundHistory[i]) / 5
	}
	energyDesat = 1 - float64((soundEnergy-float64(db.min))/float64((db.max-db.min)))
	if energyDesat > 1 {
		energyDesat = 1
	} else if energyDesat < 0 {
		energyDesat = 0
	}

	for y := 0; y < db.dataHeight; y++ {
		for x := 0; x < db.dataWidth; x++ {
			r, g, b, a := c.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 {
				clr := desaturate(energyDesat, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
				c.Set(x, y, clr)
			}
		}
	}
}

// f = 0.2 is 20% desaturation
func desaturate(f float64, clr color.RGBA) color.RGBA {
	L := 0.3*float64(clr.R) + 0.6*float64(clr.G) + 0.1*float64(clr.B)
	rx := uint8(float64(clr.R) + f*(L-float64(clr.R)))
	gx := uint8(float64(clr.G) + f*(L-float64(clr.G)))
	bx := uint8(float64(clr.B) + f*(L-float64(clr.B)))
	return color.RGBA{rx, gx, bx, 0xff}
}
