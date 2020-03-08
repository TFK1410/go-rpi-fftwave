package main

import (
	"image/color"
	"log"
	"sync"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

func initDMX(slaveAddress byte, clr *color.RGBA, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()
	err := embd.InitI2C()
	if err != nil {
		log.Fatalln(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)
	defer bus.Close()

	clr.A = 0xff

	for {
		bytes, err := bus.ReadBytes(slaveAddress, 4)
		if err != nil {
			log.Println(err)
			return
		}
		if len(bytes) > 0 && bytes[0] > 0 {
			clr.R = bytes[1]
			clr.G = bytes[2]
			clr.B = bytes[3]
		}

		select {
		case <-quit:
			log.Println("Stopping dmx reader thread")
			return
		}
	}
}
