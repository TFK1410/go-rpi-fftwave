package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sync"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/lyrics"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

func initFFTSmooth(c *rgbmatrix.Canvas, wavechan <-chan drawloops.Wave, fftOutChan <-chan []float64, dmxColor *color.RGBA, wg *sync.WaitGroup, quit <-chan struct{}) {
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

	// Setup the sound energy buffers and timers
	soundEnergyColors := make([]color.RGBA, cfg.SoundEnergy.HistoryCount)
	hueTime := time.Duration(cfg.SoundEnergy.HueTime * float64(time.Second))
	var soundEnergy float64
	var soundEnergyTimer time.Duration

	// Wait for the first wave display type to be selected
	var wave drawloops.Wave
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

		// Calculate the smoothed FFT values and the sound energy
		soundEnergy = 0
		for i := range smoothFFT {
			smoothFFT[i] = cfg.Display.FFTSmoothCurve*smoothFFT[i] + (1-cfg.Display.FFTSmoothCurve)*curFFT[i]
			soundEnergy += math.Pow(smoothFFT[i], 2)
		}

		elapsed = time.Since(start)
		start = time.Now()

		// Add the current sound energy to the history buffer and convert it into color
		soundEnergy = math.Sqrt(soundEnergy)
		copy(soundEnergyColors[1:], soundEnergyColors[0:len(soundEnergyColors)-2])
		soundEnergyColors[0] = soundHue(&soundEnergyTimer, hueTime, elapsed, soundEnergy)

		// Calculate the current state of the white dots
		whiteDotCalc(dotsValue, dotsHangTime, dotsTimeLeft, smoothFFT, elapsed)

		// Generate the current canvas to be displayed
		wave.Draw(c, *dmxColor, smoothFFT, dotsValue, soundEnergyColors)

		if dmxColor.A == 255 {
			draw.Draw(c, lyrics.LyricsOverlay.Bounds(), lyrics.LyricsOverlay, image.Point{0, 0}, draw.Over)
		}

		// Call the main render of the canvas
		c.Render()

		// fmt.Printf("Elapsed time: %v\tSound Energy: %.2f\n", elapsed, soundEnergy)
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

// Based on time the sound energy values are being translated from HSV to RGB values
func soundHue(timer *time.Duration, hueTime, elapsed time.Duration, soundEnergy float64) color.RGBA {
	var H, S, V float64
	if *timer += elapsed; *timer > hueTime {
		*timer = 0
	}

	H = timer.Seconds() / hueTime.Seconds()
	S = float64(cfg.SoundEnergy.Saturation / 100)
	V = (soundEnergy - cfg.SoundEnergy.Min) / (cfg.SoundEnergy.Max - cfg.SoundEnergy.Min)

	if V < 0 {
		V = 0
	} else if V > 1 {
		V = 1
	}

	return hsv2RGB(H, S, V)

}

// H S V parameters are all between 0 and 1
func hsv2RGB(H, S, V float64) color.RGBA {
	var r, g, b float64
	if S == 0 {
		r = V * 255
		g = V * 255
		b = V * 255
	} else {
		h := H * 6
		if h == 6 {
			h = 0
		}
		i := math.Floor(h)
		v1 := V * (1 - S)
		v2 := V * (1 - S*(h-i))
		v3 := V * (1 - S*(1-(h-i)))

		if i == 0 {
			r = V
			g = v3
			b = v1
		} else if i == 1 {
			r = v2
			g = V
			b = v1
		} else if i == 2 {
			r = v1
			g = V
			b = v3
		} else if i == 3 {
			r = v1
			g = v2
			b = V
		} else if i == 4 {
			r = v3
			g = v1
			b = V
		} else {
			r = V
			g = v1
			b = v2
		}
		// RGB results from 0 to 255
		r = r * 255
		g = g * 255
		b = b * 255
	}
	out := color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
	return out
}
