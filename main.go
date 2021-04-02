package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/lyrics"
	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

// SoundSync struct contains variables used for the synchronization of the recording and FFT threads
type SoundSync struct {
	sb   chan *soundbuffer.SoundBuffer
	wg   *sync.WaitGroup
	quit <-chan struct{}
}

var configPath = flag.String("config", "config.yml", "Path to the script configuration file")

func main() {
	// Parsing the single configuration flag
	flag.StringVar(configPath, "c", "config.yml", "Path to the script configuration file")
	flag.Parse()

	// Loading the configuration from the config path
	err := loadConfig(&cfg, *configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Take over kill signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Setup a waitGroup and quit channels for the goroutines
	var quits []chan struct{}
	var wg sync.WaitGroup

	var ss SoundSync
	sb := make(chan *soundbuffer.SoundBuffer)
	ss.sb = sb
	ss.wg = &wg

	// Setup recording buffer and start the goroutine
	r, _ := soundbuffer.NewBuffer(1 << cfg.FFT.ChunkPower)
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	go initRecord(r, cfg.SampleRate/cfg.FFT.FFTUpdateRate, ss)

	// Setup FFT thread
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	fftOutChan := make(chan []float64)
	go initFFT(1<<cfg.FFT.ChunkPower, fftOutChan, ss)

	// Initialize all the possible wave types
	drawloops.InitWaves(cfg.FFT.BinCount, cfg.Display.MinVal, cfg.Display.MaxVal)

	// Initialize the LED matrix and the canvas that goes along with it
	m, err := rgbmatrix.NewRGBLedMatrix(cfg.Matrix)
	if err != nil {
		log.Fatal(err)
	}

	c := rgbmatrix.NewCanvas(m)
	defer c.Close()

	// Setup lyrics thread
	lyricsMeasures := make(chan int)
	lyricsID := make(chan int)
	quits = addThread(&wg, quits)
	o := lyrics.Overlay{
		SizeX:       cfg.IntMatrix.Cols * 2,
		SizeY:       cfg.IntMatrix.Rows * 2,
		RefreshRate: cfg.Lyrics.RefreshRate,
		FontFile:    cfg.Lyrics.FontFile,
		LyricsDir:   cfg.Lyrics.LyricsDir,
	}

	go o.InitLyricsThread(lyricsMeasures, lyricsID, &wg, quits[len(quits)-1])

	// Setup FFT smoothing thread
	var dmxData DMXData
	waveChan := make(chan drawloops.Wave)
	quits = addThread(&wg, quits)
	go initFFTSmooth(c, waveChan, fftOutChan, &dmxData.Color, &wg, quits[len(quits)-1])
	waveChan <- drawloops.GetFirstWave()

	// Start encoder thread
	encMessage := make(chan EncoderMessage)
	quits = addThread(&wg, quits)
	go initEncoder(cfg.Encoder.DTPin, cfg.Encoder.CLKPin, cfg.Encoder.SWPin, cfg.Encoder.LongPressTime, encMessage, &wg, quits[len(quits)-1])

	// Start DMX reader thread
	quits = addThread(&wg, quits)
	pause := make(chan struct{})
	play := make(chan struct{})
	go initDMX(cfg.DMX.SlaveAddress, &dmxData, lyricsMeasures, lyricsID, &wg, quits[len(quits)-1], pause, play)

	for {
		select {
		// Handling the encoder messages
		case msg := <-encMessage:
			switch msg {
			case BrightnessUp:
				// Brightness increase
				if bright := m.GetBrightness(); bright < 100 {
					m.SetBrightness(bright + 1)
				}
			case BrightnessDown:
				// Brightness decrease
				if bright := m.GetBrightness(); bright > 0 {
					m.SetBrightness(bright - 1)
				}
			case ButtonPress:
				// This will select the next display wave pattern
				waveChan <- drawloops.GetNextWave()
			case LongPress:
				// This will toggle the DMX color display mode
				// The A value of the color is also used by other goroutines to indicate whether the DMX reading is on
				if dmxData.Color.A > 0 {
					dmxData.Color.A = 0
					pause <- struct{}{}
				} else {
					dmxData.Color.A = 255
					play <- struct{}{}
				}
			}
		// Handle the quit message by forwarding the terminate signal to all goroutines
		case <-quit:
			log.Println("Terminating goroutines")
			log.Println(len(quits), "threads to close")

			for i := range quits {
				close(quits[i])
			}

			ss.wg.Wait()
			close(encMessage)
			close(waveChan)
			close(lyricsMeasures)
			close(lyricsID)
			log.Println("DONE")
			return
		}
	}
}

func addThread(wg *sync.WaitGroup, quits []chan struct{}) []chan struct{} {
	wg.Add(1)
	return append(quits, make(chan struct{}))
}
