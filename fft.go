package main

import (
	"log"
	"time"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"
	"github.com/cpmech/gosl/fun/fftw"
)

func initFFT(bfz int, ss SoundSync) error {
	defer ss.wg.Done()

	//Generate a plan for FFTW
	var r *soundbuffer.SoundBuffer
	var data []int16
	compData := make([]complex128, bfz)
	plan := fftw.NewPlan1d(compData, false, true)

	for {
		select {
		case <-ss.quit:
			log.Println("Stopping FFT thread")
			return nil
		case r = <-ss.sb:
		}

		start := time.Now()

		data = r.Sound()
		for i := range data {
			compData[i] = complex(float64(data[i]), 0)
		}
		plan.Execute()

		elapsed := time.Since(start)
		log.Printf("Execution time: %v\n", elapsed)
	}
}
