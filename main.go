package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/TFK1410/go-rpi-fftwave/drawloops"
	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

var (
	rows                   = flag.Int("led-rows", 32, "number of rows supported")
	cols                   = flag.Int("led-cols", 64, "number of columns supported")
	parallel               = flag.Int("led-parallel", 1, "number of daisy-chained panels")
	chain                  = flag.Int("led-chain", 4, "number of displays daisy-chained")
	brightness             = flag.Int("brightness", 30, "brightness (0-100)")
	hardwareMapping        = flag.String("led-gpio-mapping", "regular", "Name of GPIO mapping used.")
	showRefresh            = flag.Bool("led-show-refresh", false, "Show refresh rate.")
	inverseColors          = flag.Bool("led-inverse", false, "Switch if your matrix has inverse colors on.")
	disableHardwarePulsing = flag.Bool("led-no-hardware-pulse", false, "Don't use hardware pin-pulse generation.")
)

//SoundSync struct contains variables used for the synchronization of the recording and FFT threads
type SoundSync struct {
	sb   chan *soundbuffer.SoundBuffer
	wg   *sync.WaitGroup
	quit <-chan struct{}
}

func main() {
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
	r, _ := soundbuffer.NewBuffer(1 << chunkPower)
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	go initRecord(r, sampleRate/fftUpdateRate, ss)

	//Setup FFT thread
	quits = addThread(&wg, quits)
	ss.quit = quits[len(quits)-1]
	fftOutChan := make(chan []float64)
	go initFFT(1<<chunkPower, fftOutChan, ss)

	drawloops.InitWaves(dataWidth, minVal, maxVal)

	config := &rgbmatrix.DefaultConfig
	config.Rows = *rows
	config.Cols = *cols
	config.Parallel = *parallel
	config.ChainLength = *chain
	config.Brightness = *brightness
	config.HardwareMapping = *hardwareMapping
	config.ShowRefreshRate = *showRefresh
	config.InverseColors = *inverseColors
	config.DisableHardwarePulsing = *disableHardwarePulsing

	m, err := rgbmatrix.NewRGBLedMatrix(config)
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
	go initEncoder(dtPin, clkPin, swPin, encMessage, &wg, quits[len(quits)-1])

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

func init() {
	flag.Parse()
}
