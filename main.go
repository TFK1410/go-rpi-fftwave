package main

import (
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

//SoundSync struct contains variables used for the synchronization of the recording and FFT threads
type SoundSync struct {
	sb   chan *soundbuffer.SoundBuffer
	wg   *sync.WaitGroup
	quit <-chan struct{}
}

func main() {
	err := loadConfig(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	//Take over kill signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	//Setup a waitGroup and quit channels for the goroutines
	var quits []chan struct{}
	var wg sync.WaitGroup

	var ss SoundSync
	sb := make(chan *soundbuffer.SoundBuffer)
	ss.sb = sb
	ss.wg = &wg

	//Setup recording buffer and start the goroutine
	r, _ := soundbuffer.NewBuffer(1 << cfg.FFT.ChunkPower)
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	go initRecord(r, cfg.SampleRate/cfg.FFT.FFTUpdateRate, ss)

	//Setup FFT thread
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	fftOutChan := make(chan []float64)
	go initFFT(1<<cfg.FFT.ChunkPower, fftOutChan, ss)

	drawloops.InitWaves(cfg.FFT.BinCount, cfg.Display.MinVal, cfg.Display.MaxVal)

	m, err := rgbmatrix.NewRGBLedMatrix(cfg.Matrix)
	if err != nil {
		log.Fatal(err)
	}

	c := rgbmatrix.NewCanvas(m)
	defer c.Close()

	//Setup FFT smoothing thread
	quits = addThread(&wg, quits)
	go initFFTSmooth(c, fftOutChan, &wg, quits[len(quits)-1])

	//Start encoder thread
	encMessage := make(chan EncoderMessage)
	quits = addThread(&wg, quits)
	go initEncoder(cfg.Encoder.DTPin, cfg.Encoder.CLKPin, cfg.Encoder.SWPin, cfg.Encoder.LongPressTime, encMessage, &wg, quits[len(quits)-1])

	for {
		select {
		case msg := <-encMessage:
			switch msg {
			case BrightnessUp:
				if bright := m.GetBrightness(); bright < 100 {
					m.SetBrightness(bright + 1)
				}
			case BrightnessDown:
				if bright := m.GetBrightness(); bright > 0 {
					m.SetBrightness(bright - 1)
				}
			case ButtonPress:
			case LongPress:
			}
		case <-quit:
			log.Println("Terminating goroutines")

			for i := range quits {
				close(quits[i])
			}

			ss.wg.Wait()
			close(encMessage)
			log.Println("DONE")
			return
		}
	}
}

func addThread(wg *sync.WaitGroup, quits []chan struct{}) []chan struct{} {
	wg.Add(1)
	return append(quits, make(chan struct{}))
}
