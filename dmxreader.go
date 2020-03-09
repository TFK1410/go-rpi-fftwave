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

	// Initialize the I2C communication using the embd package
	err := embd.InitI2C()
	if err != nil {
		log.Fatalln(err)
	}
	defer embd.CloseI2C()

	// Create a new I2CBus
	bus := embd.NewI2CBus(1)
	defer bus.Close()

	for {
		// Listen in on the I2CBus with the specified slave address
		// first four bytes are read
		// if the first one is not zero then a new color is being registered from the next 4 bytes
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
			// Close the goroutine letting the defers trigger
			log.Println("Stopping dmx reader thread")
			return
		}
	}
}
