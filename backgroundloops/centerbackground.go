package backgroundloops

import (
	"image/color"
	"math"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// CenterBackground defines the values used for the display of the wave that are specific to this pattern type
type CenterBackground struct {
	dataWidth, dataHeight int
	min, max              float64
	hueRotation           time.Duration
	timeIncrementSpan     time.Duration
	radiusIndexes         [][]int
	delayIndexes          []int
	centerX, centerY      float64
}

// InitBackgroundLoop does the initial calculation of the reused variables in the draw loop
func (cb *CenterBackground) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
	cb.dataWidth = displayWidth
	cb.dataHeight = displayHeight
	cb.min = minVal
	cb.max = maxVal
	if cb.timeIncrementSpan == 0 {
		cb.timeIncrementSpan = 20 * time.Millisecond
	}
	if cb.hueRotation == 0 {
		cb.hueRotation = 10 * time.Second
	}
	if cb.centerX == 0 && cb.centerY == 0 {
		cb.centerX = float64(displayWidth)/2 - 0.5
		cb.centerY = float64(displayHeight)/2 - 0.5
	}
	cb.radiusIndexes = calculateDistance(displayWidth, displayHeight, cb.centerX, cb.centerY)

	// get max radiusIndex
	mx := 0
	for x := 0; x < displayWidth; x++ {
		for y := 0; y < displayHeight; y++ {
			if cb.radiusIndexes[x][y] > mx {
				mx = cb.radiusIndexes[x][y]
			}
		}
	}
	cb.delayIndexes = make([]int, mx)
}

// Draw adds the background details to the canvas on the matrix
func (cb *CenterBackground) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
	var H, S, V, soundEnergy float64
	var clr color.RGBA
	H = float64(time.Now().UnixMilli()%cb.hueRotation.Milliseconds()) / float64(cb.hueRotation.Milliseconds())
	S = 1
	cb.updateDelays(soundHistory)

	for y := 0; y < cb.dataHeight; y++ {
		for x := 0; x < cb.dataWidth; x++ {
			r, g, b, a := c.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 && a == 0 {
				soundEnergy = maxTriBand(soundHistory[cb.delayIndexes[cb.radiusIndexes[x][y]-1]])
				V = (soundEnergy - cb.min) / (cb.max - cb.min)

				if V < 0 {
					continue
				} else if V > 1 {
					V = 1
				}
				V = V / 3

				clr = hsv2RGB(H, S, V)

				c.Set(x, y, clr)
			}
		}
	}
}

// updateDelays finds the closest soundenergy datapoint to the time marker provided
func (cb *CenterBackground) updateDelays(soundHistory []SoundEnergyTriBand) {
	timeNow := time.Now()
	sI := 0
	for i := 0; i < len(cb.delayIndexes); i++ {
		// find the closest time point relative to the current time to display
		curTime := timeNow.Add(-time.Duration(float64(i) * float64(cb.timeIncrementSpan)))
		for sI < len(soundHistory)-1 && math.Abs(float64(curTime.Sub(soundHistory[sI].Tm))) > math.Abs(float64(curTime.Sub(soundHistory[sI+1].Tm))) {
			sI++
		}
		cb.delayIndexes[i] = sI
	}
}
