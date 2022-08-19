package dmx

import (
	"image/color"
	"log"
	"math"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
)

//DMXData is a struct containing all the data received via DMX
type DMXData struct {
	DisplayMode        byte
	BackgroundMode     byte
	WhiteDots          bool
	ColorPalette       byte
	PaletteAngle       byte
	PalettePhaseOffset byte
	Color              color.RGBA
	LyricsDMXInfo      uint
}

func InitDMX(slaveAddress byte, data *DMXData, lyricsDMXInfo chan<- uint, wg *sync.WaitGroup, quit, pause, play <-chan struct{}) {
	defer wg.Done()

	// Initialize the I2C communication using the periph package
	bus, err := i2creg.Open("1")
	if err != nil {
		log.Fatalln(err)
	}
	defer bus.Close()

	dev := i2c.Dev{Bus: bus, Addr: uint16(slaveAddress)}
	bytes := make([]byte, 13)

	data.WhiteDots = true

	// Wait for the first signal to start the goroutine
	select {
	case <-play:
	case <-quit:
		// Close the goroutine letting the defers trigger
		log.Println("Stopping dmx reader thread")
		return
	}

	var incomingLyricData uint

	for {
		// Listen in on the I2CBus with the specified slave address, Tx is called with empty tx buffer to just receive
		err = dev.Tx([]byte{}, bytes)
		if err != nil {
			log.Println(err)
			continue
		}

		// if the first byte is not zero then new dmx data is being registered from the next 12 bytes
		// this is closely coupled with the Arduino sketch that's bundled with this code
		if len(bytes) > 0 && bytes[0] > 0 {

			data.DisplayMode = bytes[1] & 0x7            //xxxxx000
			data.BackgroundMode = (bytes[1] & 0x70) >> 4 //x000xxxx
			data.WhiteDots = (bytes[1] & 0x80) == 0
			data.ColorPalette = bytes[2] >> 2
			data.PaletteAngle = bytes[3]
			data.PalettePhaseOffset = bytes[4]

			brightness := float64(bytes[8]) / 255.0
			data.Color.R = uint8(math.Round(float64(bytes[5]) * brightness))
			data.Color.G = uint8(math.Round(float64(bytes[6]) * brightness))
			data.Color.B = uint8(math.Round(float64(bytes[7]) * brightness))

			if data.Color.R > 0 || data.Color.G > 0 || data.Color.B > 0 {
				data.Color.A = 255
			} else {
				data.Color.A = 0
			}

			// 3 bytes lyricID + 1 byte lyricProgress
			// if checks if the bytes from the MSB and below are not zeros
			if !((bytes[9] > 0 && bytes[10] == 0 && bytes[11] == 0) || (bytes[10] > 0 && bytes[11] == 0)) {
				incomingLyricData = uint(bytes[9])<<24 + uint(bytes[10])<<16 + uint(bytes[11])<<8 + uint(bytes[12])
			}

			if incomingLyricData != data.LyricsDMXInfo {
				data.LyricsDMXInfo = incomingLyricData
				// Send the new data to the lyrics goroutine without blocking
				select {
				case lyricsDMXInfo <- incomingLyricData:
				default:
				}
			}
		}

		// Enable pausing of the reader goroutine
		select {
		// case <-pause:
		// 	select {
		// 	case <-play:
		// 	case <-quit:
		// 		// Close the goroutine letting the defers trigger
		// 		log.Println("Stopping dmx reader thread")
		// 		return
		// 	}
		case <-quit:
			// Close the goroutine letting the defers trigger
			log.Println("Stopping dmx reader thread")
			return
		default:
		}
	}
}

//ResetDMX function resets the DMX values to the defaults
func ResetDMX(data *DMXData) {
	data.DisplayMode = 0
	data.BackgroundMode = 0
	data.WhiteDots = true
	data.ColorPalette = 0
	data.PaletteAngle = 0
	data.PalettePhaseOffset = 0
	data.Color.R = 0
	data.Color.G = 0
	data.Color.B = 0
	data.Color.A = 0
	data.LyricsDMXInfo = 0
}
