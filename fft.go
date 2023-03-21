package main

import (
	"log"
	"math"
	"math/cmplx"
	"os"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	"github.com/cpmech/gosl/fun/fftw"
)

const SoundEmulatorENV = "SOUND_EMULATOR"

// initFFT function is a start for the goroutine handling the FFT part of the application.
// bfz is the number of elements in a single FFT call.
// Should be the same as the ring buffer size.
func initFFT(bfz int, fftOutChan chan<- []float64, ss SoundSync) error {
	defer ss.wg.Done()
	//  start := time.Now()

	// Generate a plan for FFTW
	var r *soundbuffer.SoundBuffer
	var data []int16
	compData := make([]complex128, bfz)
	realData := make([]float64, bfz)
	plan := fftw.NewPlan1d(compData, false, true)
	defer plan.Free()

	// Calculate the logarithmic bins
	fftBins, fftBinFloating := calculateBins(cfg.Display.MinHz, cfg.Display.MaxHz, cfg.FFT.BinCount, cfg.SampleRate, 1<<cfg.FFT.ChunkPower)

	outFFT := make([]float64, cfg.FFT.BinCount)

	freq := 10

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

		// Convert int16 data into complex128
		data = r.Sound()

		if os.Getenv(SoundEmulatorENV) == "1" {
			freq = int(math.Round(float64(freq) * 1.1))
			if freq > cfg.SampleRate {
				freq = 10
			}
			for i := range data {
				data[i] = int16(freq*i*0xffff/cfg.SampleRate - 0x7fff)
			}
		}

		for i := range data {
			compData[i] = complex(float64(data[i]), 0)
		}

		// Execute the plan
		plan.Execute()

		// Convert the data to real values
		for i := range compData {
			realData[i] = cmplx.Abs(compData[i])
		}

		// Convert the linear data to logarithmic space
		fftToBins(fftBins, fftBinFloating, realData, outFFT)

		// Send the new data to the smoothing goroutine without blocking
		select {
		case fftOutChan <- outFFT:
		default:
		}

		// elapsed = time.Since(start)
		// log.Printf("Execution time: %v\n", elapsed)
	}
}

// fftToBins translates the result of Fourier transform which is in linear bins
// to bins in logarithmic space
func fftToBins(fftBins []int, fftBinFloating []float64, data, out []float64) {
	var maxFromBins, logFromMax float64
	for i := 0; i < len(out); i++ {
		if fftBins[i+1]-fftBins[i] <= 1 {
			lbin, lfrac := math.Modf(fftBinFloating[i])
			maxFromBins = math.Abs(data[int(lbin)]*(1-lfrac) + data[int(lbin)+1]*lfrac)
		} else {
			maxFromBins = maxFromRange(fftBins[i], fftBins[i+1], data)
		}

		if maxFromBins > 0 {
			logFromMax = math.Log10(maxFromBins)
		}

		out[i] = 20 * logFromMax
	}
}
