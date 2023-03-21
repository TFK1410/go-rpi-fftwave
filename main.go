package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TFK1410/go-rpi-fftwave/backgroundloops"
	"github.com/TFK1410/go-rpi-fftwave/dmx"
	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/lyricsoverlay"
	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
	"periph.io/x/host/v3"
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

	// Load GPIO and I2C drivers
	_, err = host.Init()
	if err != nil {
		log.Fatal(err)
	}

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
	backgroundloops.InitBackgroundLoops(cfg.Matrix.Cols*2, cfg.Matrix.Rows*2, cfg.SoundEnergy.MinBand, cfg.SoundEnergy.MaxBand)

	// Initialize the LED matrix and the canvas that goes along with it
	// set export MATRIX_TERMINAL_EMULATOR=1 to use the terminal emulator version for testing
	// set export SOUND_EMULATOR=1 to add dummy sound data for testing
	m, err := rgbmatrix.NewRGBLedMatrix(cfg.Matrix)
	if err != nil {
		log.Fatal(err)
	}

	c := rgbmatrix.NewCanvas(m)
	defer c.Close()

	log.Println("Canvas size:", c.Bounds().Dx(), "x", c.Bounds().Dy())

	// Setup lyrics thread
	lyricsDMXInfo := make(chan uint)
	quits = addThread(&wg, quits)
	ldc := lyricsoverlay.LyricDrawContext{
		SizeX:       c.Bounds().Dx(),
		SizeY:       c.Bounds().Dy(),
		RefreshRate: cfg.Lyrics.RefreshRate,
		SqlitePath:  cfg.Lyrics.SqlitePath,
	}

	go ldc.InitLyricsThread(lyricsDMXInfo, &wg, quits[len(quits)-1])

	// Setup FFT smoothing thread
	var dmxData dmx.DMXData
	waveChan := make(chan drawloops.Wave)
	backgroundChan := make(chan backgroundloops.BackgroundLoop)
	quits = addThread(&wg, quits)
	go initFFTSmooth(c, waveChan, backgroundChan, fftOutChan, &dmxData, &ldc, &wg, quits[len(quits)-1])
	waveChan <- drawloops.GetFirstWave()
	backgroundChan <- backgroundloops.GetFirstBackgroundLoop()

	// Start encoder thread
	encMessage := make(chan EncoderMessage)
	quits = addThread(&wg, quits)
	go initEncoder(cfg.Encoder.DTPin, cfg.Encoder.CLKPin, cfg.Encoder.SWPin, cfg.Encoder.LongPressTime, encMessage, &wg, quits[len(quits)-1])

	// Start DMX reader thread
	quits = addThread(&wg, quits)
	pause := make(chan struct{})
	play := make(chan struct{})
	go dmx.InitDMX(cfg.DMX.SlaveAddress, &dmxData, lyricsDMXInfo, &wg, quits[len(quits)-1], pause, play)
	dmx.ResetDMX(&dmxData)
	play <- struct{}{}

	log.Println("All initialized")

	for {
		select {
		// Handling the encoder messages
		case msg := <-encMessage:
			switch msg {
			case BrightnessUp:
				// Brightness increase
				if bright := m.GetBrightness(); bright < 70 {
					m.SetBrightness(bright + 1)
				}
			case BrightnessDown:
				// Brightness decrease
				if bright := m.GetBrightness(); bright > 0 {
					m.SetBrightness(bright - 1)
				}
			case ButtonPress:
				// This will select the next display wave pattern
				// if !dmxData.DMXOn {
				// 	waveChan <- drawloops.GetNextWave()
				// }
			case LongPress:
				// This will toggle the DMX color display mode
				// if dmxData.DMXOn {
				// 	pause <- struct{}{}
				// } else {
				// 	play <- struct{}{}
				// }
				// This will reset the current DMX data
				dmx.ResetDMX(&dmxData)
			case UpPress:
				waveChan <- drawloops.GetNextWave()
			case DownPress:
				backgroundChan <- backgroundloops.GetNextBackgroundLoop()

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
			close(lyricsDMXInfo)
			log.Println("DONE")
			return
		}
	}
}

func addThread(wg *sync.WaitGroup, quits []chan struct{}) []chan struct{} {
	wg.Add(1)
	return append(quits, make(chan struct{}))
}
