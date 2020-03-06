package main

import (
	"fmt"
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
	encoderPush       int
	encoderState      int
	pressTimer        time.Time
)

func initEncoder(DTpin, CLKpin, SWpin int, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()
	embd.InitGPIO()
	defer embd.CloseGPIO()

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
		encoderPush = 1
		fmt.Println("Pushed encoder")
		pressTimer = time.Now()
	} else if currentRoSWStatus == 1 && lastRoSWStatus == 0 {
		encoderPush = 0
		fmt.Println("Let go of encoder")
		if time.Since(pressTimer) > longPress {
			fmt.Println("Long pressed")
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
		encoderState++
	} else if lastRoBStatus == 1 && currentRoBStatus == 0 {
		encoderState--
	}

	fmt.Println("Rotated encoder:", encoderState)
}

func nonBlockingChannelSend(msg int, c chan<- int) {
	select {
	case c <- msg:
		//message sent
	default:
		//message dropped
	}
}
