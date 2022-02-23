package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

// Set of global variables used throughout
var (
	roDTPin            gpio.PinIO
	roCLKPin           gpio.PinIO
	roSWPin            gpio.PinIO
	currentRoCLKStatus gpio.Level
	lastRoCLKStatus    gpio.Level
	currentRoSWStatus  gpio.Level
	lastRoSWStatus     gpio.Level
	pressTimer         time.Time
	longPressTime      time.Duration
	rotateTimer        time.Time
	rotateDelay        time.Duration
	encoderChannel     chan<- EncoderMessage
)

// EncoderMessage defines the kinds of messages that can originate from encoder actions
type EncoderMessage int

// Definition of message types
const (
	BrightnessUp EncoderMessage = iota
	BrightnessDown
	ButtonPress
	LongPress
	UpPress
	DownPress
)

func initEncoder(DTpin, CLKpin, SWpin int, pressTime float64, messages chan<- EncoderMessage, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()

	encoderChannel = messages
	longPressTime = time.Duration(pressTime * float64(time.Second))

	// Debounce rotations, limit triggers to once every 100 microseconds
	rotateDelay = time.Duration(100 * float64(time.Microsecond))

	// Init the DTPin
	roDTPin = gpioreg.ByName(fmt.Sprint(DTpin))
	if roDTPin == nil {
		log.Fatalln("Failed to find " + fmt.Sprint(DTpin))
	}

	// Init the CLKPin
	roCLKPin = gpioreg.ByName(fmt.Sprint(CLKpin))
	if roCLKPin == nil {
		log.Fatalln("Failed to find " + fmt.Sprint(CLKpin))
	}

	// Init the SWPin
	roSWPin = gpioreg.ByName(fmt.Sprint(SWpin))
	if roSWPin == nil {
		log.Fatalln("Failed to find " + fmt.Sprint(SWpin))
	}

	// Set all as input, with an internal pull down resistor:
	if err := roDTPin.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatalln(err)
	}
	if err := roCLKPin.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatalln(err)
	}
	if err := roSWPin.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatalln(err)
	}

	// Set the variable below so that the first press is properly triggered
	lastRoSWStatus = gpio.High
	lastRoCLKStatus = gpio.High

	// Setup callback functions for the pins
	go func() {
		for {
			roSWPin.WaitForEdge(-1)
			callPress()
		}
	}()

	go func() {
		for {
			roDTPin.WaitForEdge(-1)
			callRotate()
		}
	}()

	<-quit
	// Stop the thread and let the defers trigger
	log.Println("Stopping encoder thread")
}

var pressNoMove, pressed bool

// Called when the button is pressed
// debouncing should be taken care of over here
// short press and long press send two different messages back to the main function
func callPress() {
	currentRoSWStatus = roSWPin.Read()
	if currentRoSWStatus == gpio.Low && lastRoSWStatus == gpio.High {
		pressTimer = time.Now()
		pressNoMove = true
		pressed = true
	} else if currentRoSWStatus == gpio.High && lastRoSWStatus == gpio.Low && pressNoMove {
		if time.Since(pressTimer) > longPressTime {
			sendMessage(LongPress, encoderChannel)
		} else {
			sendMessage(ButtonPress, encoderChannel)
		}
	}
	lastRoSWStatus = currentRoSWStatus
}

// Called when the encoder is rotated
func callRotate() {
	if pinVal := roDTPin.Read(); pinVal == gpio.High {
		lastRoCLKStatus = roCLKPin.Read()
	} else {
		currentRoCLKStatus = roCLKPin.Read()
		rotateTimer = time.Now()
	}

	if time.Since(rotateTimer) > rotateDelay {
		if lastRoCLKStatus == gpio.High && currentRoCLKStatus == gpio.Low {
			if pressed {
				sendMessage(UpPress, encoderChannel)
			} else {
				sendMessage(BrightnessUp, encoderChannel)
			}
		} else if lastRoCLKStatus == gpio.Low && currentRoCLKStatus == gpio.High {
			if pressed {
				sendMessage(DownPress, encoderChannel)
			} else {
				sendMessage(BrightnessDown, encoderChannel)
			}
		}
	}
}

// This will send the message but it won't block if the channel is not listened to at this moment
func sendMessage(msg EncoderMessage, c chan<- EncoderMessage) {
	select {
	case c <- msg:
	default:
	}
}
