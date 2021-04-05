package dmx

import (
	"image/color"
	"log"
	"math"
	"sync"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

//DMXData is a struct containing all the data received via DMX
type DMXData struct {
	DMXOn              bool
	DisplayMode        byte
	WhiteDots          bool
	ColorPalette       byte
	PaletteAngle       byte
	PalettePhaseOffset byte
	Color              color.RGBA
	LyricID            int
	LyricProgress      byte
}

func InitDMX(slaveAddress byte, data *DMXData, lyricProgress chan<- byte, lyricID chan<- int, wg *sync.WaitGroup, quit, pause, play <-chan struct{}) {
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

	data.DMXOn = false

	// Wait for the first signal to start the goroutine
	select {
	case <-play:
		data.DMXOn = true
	case <-quit:
		// Close the goroutine letting the defers trigger
		log.Println("Stopping dmx reader thread")
		return
	}

	for {
		// Listen in on the I2CBus with the specified slave address
		bytes, err := bus.ReadBytes(slaveAddress, 13)
		if err != nil {
			log.Println(err)
			continue
		}

		// if the first byte is not zero then new dmx data is being registered from the next 12 bytes
		// this is closely coupled with the Arduino sketch that's bundled with this code
		if len(bytes) > 0 && bytes[0] > 0 {

			data.DisplayMode = bytes[1] & 0x7f
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

			incomingLyricID := int(bytes[9])<<16 + int(bytes[10])<<8 + int(bytes[11])
			incomingLyricProgress := bytes[12]

			if incomingLyricID != data.LyricID {
				data.LyricID = incomingLyricID
				// Send the new data to the lyrics goroutine without blocking
				select {
				case lyricID <- incomingLyricID:
				default:
				}
			}

			if incomingLyricProgress != data.LyricProgress {
				data.LyricProgress = incomingLyricProgress
				// Send the new data to the lyrics goroutine without blocking
				select {
				case lyricProgress <- incomingLyricProgress:
				default:
				}
			}
		}

		// Enable pausing of the reader goroutine
		select {
		case <-pause:
			data.DMXOn = false
			select {
			case <-play:
				data.DMXOn = true
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
