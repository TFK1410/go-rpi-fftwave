package backgroundloops

import (
	"time"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

type SoundEnergyTriBand struct {
	Bass, Mid, Treble float64
	Tm                time.Time
}

// Wave is used for the implementation of any possible display patterns
type BackgroundLoop interface {
	InitBackgroundLoop(int, int, float64, float64)
	Draw(*rgbmatrix.Canvas, dmx.DMXData, []SoundEnergyTriBand)
}

var iterator int
var backgroundLoops []BackgroundLoop

// InitBackgroundLoops creates the BackgroundLoop types array and initializes every one of them
func InitBackgroundLoops(displayWidth int, displayHeight int, minVal, maxVal float64) {
	backgroundLoops = append(backgroundLoops, &CenterBackground{})
	backgroundLoops = append(backgroundLoops, &CenterBackgroundInst{})
	backgroundLoops = append(backgroundLoops, &HistoryBackground{timeSpan: 1000 * time.Millisecond})
	backgroundLoops = append(backgroundLoops, &ShiftHueBackground{})
	backgroundLoops = append(backgroundLoops, &DesaturateBackground{})
	backgroundLoops = append(backgroundLoops, &NoBackground{})

	for i := range backgroundLoops {
		backgroundLoops[i].InitBackgroundLoop(displayWidth, displayHeight, minVal, maxVal)
	}
}

// GetFirstWave returns the first BackgroundLoop type from the array
func GetFirstBackgroundLoop() BackgroundLoop {
	return backgroundLoops[0]
}

func GetBackgroundLoopNum(i int) BackgroundLoop {
	iterator = i
	if iterator >= len(backgroundLoops) {
		iterator = 0
	}
	return backgroundLoops[iterator]
}

// GetNextWave returns the next BackgroundLoop type from the slice
func GetNextBackgroundLoop() BackgroundLoop {
	iterator++
	if iterator >= len(backgroundLoops) {
		iterator = 0
	}
	return backgroundLoops[iterator]
}
