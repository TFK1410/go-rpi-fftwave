package backgroundloops

import (
	"image/color"
	"math"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// ShiftHueBackground defines the values used for the display of the wave that are specific to this pattern type
type ShiftHueBackground struct {
	dataWidth, dataHeight int
	min, max              float64
	matrix                [3][3]float64
}

// InitBackgroundLoop does the initial calculation of the reused variables in the draw loop
func (shb *ShiftHueBackground) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
	shb.dataWidth = displayWidth
	shb.dataHeight = displayHeight
	shb.min = minVal
	shb.max = maxVal
	shb.matrix = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
}

// Draw adds the background details to the canvas on the matrix
func (shb *ShiftHueBackground) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
	var soundEnergy, energyAngle float64

	soundEnergy = 0
	for i := 0; i < 5; i++ {
		soundEnergy += maxTriBand(soundHistory[i]) / 5
	}
	energyAngle = float64((soundEnergy - float64(shb.min)) / float64((shb.max - shb.min)) * 90)
	if energyAngle > 90 {
		energyAngle = 90
	} else if energyAngle < 0 {
		energyAngle = 0
	}
	shb.setHueRotation(energyAngle)

	for y := 0; y < shb.dataHeight; y++ {
		for x := 0; x < shb.dataWidth; x++ {
			r, g, b, a := c.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 {
				clr := shb.applyShift(color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
				c.Set(x, y, clr)
			}
		}
	}
}

func (shb *ShiftHueBackground) setHueRotation(degrees float64) {
	radians := degrees * math.Pi / 180
	squared := math.Sqrt(1.0 / 3.0)
	cosA := math.Cos(radians)
	sinA := math.Sin(radians)
	shb.matrix[0][0] = cosA + (1.0-cosA)/3.0
	shb.matrix[0][1] = 1./3.*(1.0-cosA) - squared*sinA
	shb.matrix[0][2] = 1./3.*(1.0-cosA) + squared*sinA
	shb.matrix[1][0] = 1./3.*(1.0-cosA) + squared*sinA
	shb.matrix[1][1] = cosA + 1./3.*(1.0-cosA)
	shb.matrix[1][2] = 1./3.*(1.0-cosA) - squared*sinA
	shb.matrix[2][0] = 1./3.*(1.0-cosA) - squared*sinA
	shb.matrix[2][1] = 1./3.*(1.0-cosA) + squared*sinA
	shb.matrix[2][2] = cosA + 1./3.*(1.0-cosA)
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func (shb *ShiftHueBackground) applyShift(clr color.RGBA) color.RGBA {
	rx := float64(clr.R)*shb.matrix[0][0] + float64(clr.G)*shb.matrix[0][1] + float64(clr.B)*shb.matrix[0][2]
	gx := float64(clr.R)*shb.matrix[1][0] + float64(clr.G)*shb.matrix[1][1] + float64(clr.B)*shb.matrix[1][2]
	bx := float64(clr.R)*shb.matrix[2][0] + float64(clr.G)*shb.matrix[2][1] + float64(clr.B)*shb.matrix[2][2]
	return color.RGBA{clamp(rx), clamp(gx), clamp(bx), 0xff}
}
