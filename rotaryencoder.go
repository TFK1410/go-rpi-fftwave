package main

import (
	"log"
	"sync"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

// Set of global variables used throughout
var (
	roAPin            embd.DigitalPin
	roBPin            embd.DigitalPin
	roSPin            embd.DigitalPin
	currentRoBStatus  int
	lastRoBStatus     int
	currentRoSWStatus int
	lastRoSWStatus    int
	pressTimer        time.Time
	longPressTime     time.Duration
	encoderChannel    chan<- EncoderMessage
)

// EncoderMessage defines the kinds of messages that can originate from encoder actions
type EncoderMessage int

// Definition of message types
const (
	BrightnessUp EncoderMessage = iota
	BrightnessDown
	ButtonPress
	LongPress
)

func initEncoder(DTpin, CLKpin, SWpin int, pressTime float64, messages chan<- EncoderMessage, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()
	// Initialize the GPIO functions using the embd package
	err := embd.InitGPIO()
	if err != nil {
		log.Fatalln(err)
	}
	defer embd.CloseGPIO()

	encoderChannel = messages
	longPressTime = time.Duration(pressTime * float64(time.Second))

	// Init the DTPin
	roAPin, err = embd.NewDigitalPin(DTpin)
	if err != nil {
		log.Fatalln(err)
	}

	// Init the CLKpin
	roBPin, err = embd.NewDigitalPin(CLKpin)
	if err != nil {
		log.Fatalln(err)
	}

	// Init the SWpin
	roSPin, err = embd.NewDigitalPin(SWpin)
	if err != nil {
		log.Fatalln(err)
	}

	// Set all pin directions as inputs
	roAPin.SetDirection(embd.In)
	roBPin.SetDirection(embd.In)
	roSPin.SetDirection(embd.In)

	// Set the variable below so that the first press is properly triggered
	lastRoSWStatus = 1

	// Setup callback functions for the pins
	roSPin.Watch(embd.EdgeBoth, callClear)
	roAPin.Watch(embd.EdgeBoth, callDeal)

	select {
	case <-quit:
		// Stop the thread and let the defers trigger
		log.Println("Stopping encoder thread")
		return
	}
}

// Called when the button is pressed
// debouncing should be taken care of over here
// short press and long press send two different messages back to the main function
func callClear(pin embd.DigitalPin) {
	currentRoSWStatus, _ = pin.Read()
	if currentRoSWStatus == 0 && lastRoSWStatus == 1 {
		pressTimer = time.Now()
	} else if currentRoSWStatus == 1 && lastRoSWStatus == 0 {
		if time.Since(pressTimer) > longPressTime {
			sendMessage(LongPress, encoderChannel)
		} else {
			sendMessage(ButtonPress, encoderChannel)
		}
	}
	lastRoSWStatus = currentRoSWStatus
}

// Called when the encoder is rotated
func callDeal(pin embd.DigitalPin) {
	if pinVal, _ := pin.Read(); pinVal == 0 {
		lastRoBStatus, _ = roBPin.Read()
	} else {
		currentRoBStatus, _ = roBPin.Read()
	}

	if lastRoBStatus == 0 && currentRoBStatus == 1 {
		sendMessage(BrightnessUp, encoderChannel)
	} else if lastRoBStatus == 1 && currentRoBStatus == 0 {
		sendMessage(BrightnessDown, encoderChannel)
	}
}

// This will send the message but it won't block if the channel is not listened to at this moment
func sendMessage(msg EncoderMessage, c chan<- EncoderMessage) {
	select {
	case c <- msg:
	default:
	}
}
