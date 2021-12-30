package main

import (
	"image"
	"log"
	"sync"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/backgroundloops"
	"github.com/TFK1410/go-rpi-fftwave/dmx"
	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/lyricsoverlay"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

func initFFTSmooth(c *rgbmatrix.Canvas, wavechan <-chan drawloops.Wave, fftOutChan <-chan []float64, dmxData *dmx.DMXData, ldc *lyricsoverlay.LyricDrawContext, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()

	// Wait for the first batch of FFT data
	var curFFT []float64
	select {
	case <-quit:
		return
	case curFFT = <-fftOutChan:
	}

	// Create a loop ticker that will try to keep the display in the specified refresh rate
	ticker := time.Tick(time.Second / time.Duration(cfg.Display.RefreshRate))

	// Create the buffer for the smoothed out FFT data to be displayed
	smoothFFT := make([]float64, cfg.FFT.BinCount)

	// Setup the white dot buffers and timers
	dotsValue := make([]float64, cfg.FFT.BinCount)
	dotsTimeLeft := make([]time.Duration, cfg.FFT.BinCount)
	dotsHangTime := time.Duration(cfg.WhiteDot.HangTime * float64(time.Second))
	var start time.Time
	var elapsed time.Duration

	var soundTriBandMax backgroundloops.SoundEnergyTriBand
	soundTriBandMaxHistory := make([]backgroundloops.SoundEnergyTriBand, cfg.SoundEnergy.HistoryCount)

	// Wait for the first wave display type to be selected
	var wave drawloops.Wave
	background := backgroundloops.GetFirstBackgroundLoop()
	select {
	case <-quit:
		return
	case wave = <-wavechan:
	}

	for {
		select {
		case <-quit:
			log.Println("Stopping FFT smoothing thread")
			return
		case curFFT = <-fftOutChan:
			continue
		case wave = <-wavechan:
			continue
		case <-ticker:
		}

		if dmxData.DMXOn {
			wave = drawloops.GetWaveNum(int(dmxData.DisplayMode))
			background = backgroundloops.GetBackgroundLoopNum(int(dmxData.BackgroundMode))
		}

		// Calculate the smoothed FFT values and the sound energy
		soundTriBandMax.Bass, soundTriBandMax.Mid, soundTriBandMax.Treble = 0, 0, 0
		// soundEnergy = 0
		for i := range smoothFFT {
			smoothFFT[i] = cfg.Display.FFTSmoothCurve*smoothFFT[i] + (1-cfg.Display.FFTSmoothCurve)*curFFT[i]
			bandIndex := 3 * i / len(smoothFFT)
			switch bandIndex {
			case 0:
				if soundTriBandMax.Bass < smoothFFT[i] {
					soundTriBandMax.Bass = smoothFFT[i]
				}
			case 1:
				if soundTriBandMax.Mid < smoothFFT[i] {
					soundTriBandMax.Mid = smoothFFT[i]
				}
			case 2:
				if soundTriBandMax.Treble < smoothFFT[i] {
					soundTriBandMax.Treble = smoothFFT[i]
				}
			}
		}
		// fmt.Println(soundTriBandMax.Bass, soundTriBandMax.Mid, soundTriBandMax.Treble)

		elapsed = time.Since(start)
		start = time.Now()
		soundTriBandMax.Tm = start

		// Add the current sound energy to the history buffer
		copy(soundTriBandMaxHistory[1:], soundTriBandMaxHistory[0:len(soundTriBandMaxHistory)-2])
		soundTriBandMaxHistory[0] = soundTriBandMax

		// Calculate the current state of the white dots
		whiteDotCalc(dotsValue, dotsHangTime, dotsTimeLeft, smoothFFT, elapsed)

		// Generate the current canvas to be displayed
		wave.Draw(c, *dmxData, smoothFFT, dotsValue)

		if dmxData.DMXOn {
			overlay := ldc.GetImage()
			// overlay := image.NewRGBA(image.Rect(0, 0, c.Bounds().Dx(), c.Bounds().Dy()))
			// draw.Draw(overlay, overlay.Bounds(), &image.Uniform{color.White}, image.Point{0, 0}, draw.Src)

			// starttest := time.Now()
			overlayImage(c, overlay)
			// draw.Draw(c, overlay.Bounds(), overlay, image.Point{0, 0}, draw.Over)
			// log.Println("Elapsed: ", time.Since(starttest))
		}

		background.Draw(c, *dmxData, soundTriBandMaxHistory)

		// Call the main render of the canvas
		c.Render()

		// fmt.Printf("Elapsed time: %v\tSound Energy: %.2f\n", elapsed, soundEnergy)
	}
}

func overlayImage(c *rgbmatrix.Canvas, overlay *image.RGBA) {
	bounds := overlay.Bounds()
	sizeX, sizeY := bounds.Dx(), bounds.Dy()

	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			if overlay.Pix[overlay.PixOffset(x, y)+3] > 0 {
				c.Set(x, y, overlay.At(x, y))
			}
		}
	}
}

// Calculate the elapsed time for the white dots hang and lower the values if necessary
func whiteDotCalc(dotsValue []float64, hangTime time.Duration, dotsTimeLeft []time.Duration, fft []float64, elapsed time.Duration) {
	for i := range dotsValue {
		if dotsValue[i] < fft[i] {
			dotsValue[i] = fft[i]
			dotsTimeLeft[i] = hangTime
		} else {
			if dotsTimeLeft[i] > 0 {
				dotsTimeLeft[i] -= elapsed
			}
			if dotsTimeLeft[i] <= 0 {
				dotsValue[i] -= elapsed.Seconds() * cfg.WhiteDot.DropSpeed
			}
		}
	}
}
