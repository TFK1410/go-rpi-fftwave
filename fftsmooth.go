package main

import (
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

	for {
		select {
		case <-quit:
			return
		case curFFT = <-fftOutChan:
			continue
		case <-ticker:
		}

		for i := range smoothFFT {
			smoothFFT[i] = fftSmoothCurve*smoothFFT[i] + (1-fftSmoothCurve)*curFFT[i]
		}

		drawloops.Draw(c, smoothFFT)
	}
}
