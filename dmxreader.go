package main

import (
	"image/color"
	"log"
	"sync"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

//DMXData is a struct containing all the data received via DMX
type DMXData struct {
	ID      int
	Measure int
	Color   color.RGBA
}

func initDMX(slaveAddress byte, data *DMXData, lyricsMeasure, lyricsID chan<- int, wg *sync.WaitGroup, quit, pause, play <-chan struct{}) {
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

	// Wait for the first signal to start the goroutine
	select {
	case <-play:
	case <-quit:
		// Close the goroutine letting the defers trigger
		log.Println("Stopping dmx reader thread")
		return
	}

	for {
		// Listen in on the I2CBus with the specified slave address
		bytes, err := bus.ReadBytes(slaveAddress, 8)
		if err != nil {
			log.Println(err)
			continue
		}

		// if the first byte is not zero then new dmx data is being registered from the next 7 bytes
		// this is closely coupled with the Arduino sketch that's bundled with this code
		if len(bytes) > 0 && bytes[0] > 0 {

			if bytes[1] > 0 && bytes[2] < 255 {
				incomingMeasure := int(bytes[2]-1)*255.0 + int(bytes[1])
				if incomingMeasure != data.Measure {
					// Send the new data to the lyrics goroutine without blocking
					select {
					case lyricsMeasure <- incomingMeasure:
					default:
					}
				}
				data.Measure = incomingMeasure
			}

			if bytes[3] > 0 && bytes[4] < 255 {
				incomingID := int(bytes[4])*255.0 + int(bytes[3])
				if incomingID != data.ID {
					// Send the new data to the lyrics goroutine without blocking
					select {
					case lyricsID <- incomingID:
					default:
					}
				}
				data.ID = incomingID
			}

			data.Color.R = bytes[5]
			data.Color.G = bytes[6]
			data.Color.B = bytes[7]
		}

		// Enable pausing of the reader goroutine
		select {
		case <-pause:
			select {
			case <-play:
			case <-quit:
				// Close the goroutine letting the defers trigger
				log.Println("Stopping dmx reader thread")
				return
			}
		case <-quit:
			// Close the goroutine letting the defers trigger
			log.Println("Stopping dmx reader thread")
			return
		default:
		}
	}
}
