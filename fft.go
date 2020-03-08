package main

import (
	"log"
	"math"
	"math/cmplx"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	"github.com/cpmech/gosl/fun/fftw"
)

//initFFT function is a start for the goroutine handling the FFT part of the application.
//bfz is the number of elements in a single FFT call.
//Should be the same as the ring buffer size.
func initFFT(bfz int, fftOutChan chan<- []float64, ss SoundSync) error {
	defer ss.wg.Done()
	// start := time.Now()

	//Generate a plan for FFTW
	var r *soundbuffer.SoundBuffer
	var data []int16
	compData := make([]complex128, bfz)
	realData := make([]float64, bfz)
	plan := fftw.NewPlan1d(compData, false, true)
	defer plan.Free()

	fftBins := calculateBins(cfg.Display.MinHz, cfg.Display.MaxHz, cfg.FFT.BinCount, cfg.SampleRate, 1<<cfg.FFT.ChunkPower)

	outFFT := make([]float64, cfg.FFT.BinCount)

	for {
		select {
		case <-ss.quit:
			log.Println("Stopping FFT thread")
			return nil
		case r = <-ss.sb:
		}

		// elapsed := time.Since(start)
		// log.Printf("Sleep time: %v\n", elapsed)
		// start = time.Now()

		//Convert int16 data into complex128
		data = r.Sound()
		for i := range data {
			compData[i] = complex(float64(data[i]), 0)
		}

		//Execute the plan
		plan.Execute()

		//Convert the data to real values
		for i := range compData {
			realData[i] = cmplx.Abs(compData[i])
		}

		fftToBins(fftBins, realData, outFFT)

		select {
		case fftOutChan <- outFFT:
		default:
		}

		// elapsed = time.Since(start)
		// log.Printf("Execution time: %v\n", elapsed)
	}
}

//fftToBins translates the result of Fourier transform which is in linear bins
//to bins in logarithmic space
func fftToBins(fftBins []int, data, out []float64) {
	var maxFromBins, logFromMax float64
	for i := 0; i < len(out); i++ {
		if fftBins[i] != fftBins[i+1] {
			maxFromBins = maxFromRange(fftBins[i], fftBins[i+1], data)
		} else {
			maxFromBins = math.Abs((data[fftBins[i]] + data[fftBins[i+1]]) / 2)
		}

		if maxFromBins > 0 {
			logFromMax = math.Log10(maxFromBins)
		}

		out[i] = 20 * logFromMax
	}
}
