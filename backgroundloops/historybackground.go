package backgroundloops

import (
	"image/color"
	"math"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// HistoryBackground defines the values used for the display of the wave that are specific to this pattern type
type HistoryBackground struct {
	dataWidth, dataHeight int
	min, max, bandPoints  float64
	timeSpan              time.Duration
}

// InitBackgroundLoop does the initial calculation of the reused variables in the draw loop
func (b *HistoryBackground) InitBackgroundLoop(displayWidth int, displayHeight int, minVal, maxVal float64) {
	b.dataWidth = displayWidth
	b.dataHeight = displayHeight / 2
	b.min = minVal
	b.max = maxVal
	b.bandPoints = float64(b.dataHeight) / 3.0
	b.timeSpan = 1000 * time.Millisecond
}

// Draw adds the background details to the canvas on the matrix
func (b *HistoryBackground) Draw(c *rgbmatrix.Canvas, dmxData dmx.DMXData, soundHistory []SoundEnergyTriBand) {
	sI := 0 //soundIndex
	timeNow := time.Now()
	for i := 0; i < b.dataWidth; i++ {
		// find the closest time point relative to the current time to display
		curTime := timeNow.Add(-time.Duration(float64(i) / float64(b.dataWidth) * float64(b.timeSpan)))
		for sI < len(soundHistory)-1 && math.Abs(float64(curTime.Sub(soundHistory[sI].Tm))) > math.Abs(float64(curTime.Sub(soundHistory[sI+1].Tm))) {
			sI++
		}
		// fmt.Println(curTime.Sub(soundHistory[sI].Tm), curTime.Sub(soundHistory[sI+1].Tm), sI)
		bassPoints := math.Round(soundHistory[sI].Bass * b.bandPoints / b.max)
		midPoints := math.Round(soundHistory[sI].Mid * b.bandPoints / b.max)
		treblePoints := math.Round(soundHistory[sI].Treble * b.bandPoints / b.max)

		j := 0
		for z := 0; z < int(treblePoints); z++ {
			b.drawPixels(c, b.dataWidth-1-i, j, color.RGBA{0x40, 0x40, 0x40, 255})
			j++
		}
		for z := 0; z < int(midPoints); z++ {
			b.drawPixels(c, b.dataWidth-1-i, j, color.RGBA{0x40, 0, 0, 255})
			j++
		}
		for z := 0; z < int(bassPoints); z++ {
			b.drawPixels(c, b.dataWidth-1-i, j, color.RGBA{0, 0, 0x40, 255})
			j++
		}
	}
}

// This function will mirror out a single pixel draw to multiple fields as required
func (b *HistoryBackground) drawPixels(c *rgbmatrix.Canvas, x, y int, clr color.RGBA) {
	r, g, bl, a := c.At(x, b.dataHeight+y).RGBA()
	if r == 0 && g == 0 && bl == 0 && a == 0 {
		c.Set(x, b.dataHeight+y, clr)
	}
	r, g, bl, a = c.At(x, b.dataHeight-1-y).RGBA()
	if r == 0 && g == 0 && bl == 0 && a == 0 {
		c.Set(x, b.dataHeight-1-y, clr)
	}
}
