package main

import (
	"flag"
	"image/color"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

var (
	rows                   = flag.Int("led-rows", 32, "number of rows supported")
	cols                   = flag.Int("led-cols", 64, "number of columns supported")
	parallel               = flag.Int("led-parallel", 1, "number of daisy-chained panels")
	chain                  = flag.Int("led-chain", 4, "number of displays daisy-chained")
	brightness             = flag.Int("brightness", 100, "brightness (0-100)")
	hardwareMapping        = flag.String("led-gpio-mapping", "regular", "Name of GPIO mapping used.")
	showRefresh            = flag.Bool("led-show-refresh", false, "Show refresh rate.")
	inverseColors          = flag.Bool("led-inverse", false, "Switch if your matrix has inverse colors on.")
	disableHardwarePulsing = flag.Bool("led-no-hardware-pulse", false, "Don't use hardware pin-pulse generation.")
)

func main() {
	//Take over kill signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	//Setup a waitGroup for the goroutines
	var wg sync.WaitGroup
	wg.Add(1)

	//Setup recording buffer and start the goroutine
	recQuit := make(chan bool)
	r, _ := soundbuffer.NewBuffer(1 << chunkPower)
	go initRecord(r, sampleRate/fftUpdateRate, &wg, recQuit)

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
	fatal(err)

	c := rgbmatrix.NewCanvas(m)
	defer c.Close()

	bounds := c.Bounds()

	go func() {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				//fmt.Println("x", x, "y", y)

				c.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
			c.Render()
		}
	}()

	for {
		select {
		case <-quit:
			log.Println("CLOSING")
			recQuit <- true
			wg.Wait()
			log.Println("DONE")
			return
		}
	}
}

func init() {
	flag.Parse()
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
