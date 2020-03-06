package main

import (
	"log"
	"sync"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
)

var (
	roAPin            embd.DigitalPin
	roBPin            embd.DigitalPin
	roSPin            embd.DigitalPin
	currentRoBStatus  int
	lastRoBStatus     int
	currentRoSWStatus int
	lastRoSWStatus    int
	pressTimer        time.Time
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

func initEncoder(DTpin, CLKpin, SWpin int, messages chan<- EncoderMessage, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()
	embd.InitGPIO()
	defer embd.CloseGPIO()

	encoderChannel = messages

	var err error

	roAPin, err = embd.NewDigitalPin(DTpin)
	if err != nil {
		log.Fatalln(err)
	}

	roBPin, err = embd.NewDigitalPin(CLKpin)
	if err != nil {
		log.Fatalln(err)
	}

	roSPin, err = embd.NewDigitalPin(SWpin)
	if err != nil {
		log.Fatalln(err)
	}

	roAPin.SetDirection(embd.In)
	roBPin.SetDirection(embd.In)
	roSPin.SetDirection(embd.In)

	lastRoSWStatus = 1

	roSPin.Watch(embd.EdgeBoth, callClear)
	roAPin.Watch(embd.EdgeBoth, callDeal)

	select {
	case <-quit:
		log.Println("Stopping encoder thread")
		return
	}
}

func callClear(pin embd.DigitalPin) {
	currentRoSWStatus, _ = pin.Read()
	if currentRoSWStatus == 0 && lastRoSWStatus == 1 {
		pressTimer = time.Now()
	} else if currentRoSWStatus == 1 && lastRoSWStatus == 0 {
		if time.Since(pressTimer) > longPress {
			sendMessage(LongPress, encoderChannel)
		} else {
			sendMessage(ButtonPress, encoderChannel)
		}
	}
	lastRoSWStatus = currentRoSWStatus
}

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

func sendMessage(msg EncoderMessage, c chan<- EncoderMessage) {
	select {
	case c <- msg:
		//message sent
	default:
		//message dropped
	}
}
