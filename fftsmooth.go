package main

import (
	"math"
	"sync"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

func initFFTSmooth(c *rgbmatrix.Canvas, fftOutChan <-chan []float64, wg *sync.WaitGroup, quit <-chan bool) {
	defer wg.Done()

	//Wait for the first batch of FFT data
	var curFFT []float64
	select {
	case <-quit:
		return
	case curFFT = <-fftOutChan:
	}

	ticker := time.Tick(time.Second / targetRefreshRate)
	smoothFFT := make([]float64, dataWidth)

	dotsValue := make([]float64, dataWidth)
	dotsTimeLeft := make([]time.Duration, dataWidth)
	var start time.Time
	var elapsed time.Duration

	soundEnergyHistory := make([]float64, soundEnergyHistoryCount)
	var soundEnergy float64

	for {
		select {
		case <-quit:
			return
		case curFFT = <-fftOutChan:
			continue
		case <-ticker:
		}

		soundEnergy = 0
		for i := range smoothFFT {
			smoothFFT[i] = fftSmoothCurve*smoothFFT[i] + (1-fftSmoothCurve)*curFFT[i]
			soundEnergy += math.Pow(smoothFFT[i], 2)
		}

		copy(soundEnergyHistory[1:], soundEnergyHistory[0:len(soundEnergyHistory)-2])
		soundEnergyHistory[0] = math.Sqrt(soundEnergy)

		elapsed = time.Since(start)
		whiteDotCalc(dotsValue, dotsTimeLeft, smoothFFT, elapsed)
		start = time.Now()

		drawloops.BasicWave.Draw(c, smoothFFT, dotsValue, soundEnergyHistory)
		//fmt.Printf("Elapsed time: %v\tSound Energy: %.2f\n", elapsed, soundEnergyHistory[0])
	}
}

func whiteDotCalc(dotsValue []float64, dotsTimeLeft []time.Duration, fft []float64, elapsed time.Duration) {
	for i := range dotsValue {
		if dotsValue[i] < fft[i] {
			dotsValue[i] = fft[i]
			dotsTimeLeft[i] = whiteDotHangTime
		} else {
			if dotsTimeLeft[i] > 0 {
				dotsTimeLeft[i] -= elapsed
			}
			if dotsTimeLeft[i] <= 0 {
				dotsValue[i] -= elapsed.Seconds() * whiteDotDropSpeed
			}
		}
	}
}
