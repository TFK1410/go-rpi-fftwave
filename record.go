package main

import (
	"encoding/binary"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"

	"github.com/gordonklaus/portaudio"
)

func initRecord(r *soundbuffer.SoundBuffer, samplesPerFrame int, wg *sync.WaitGroup, quit <-chan bool) error {
	defer wg.Done()

	log.Println("Setting up signal handling for recording")
	record := make(chan os.Signal, 1)
	signal.Notify(record, syscall.SIGUSR1)

	log.Println("Initializing PortAudio")
	portaudio.Initialize()
	defer portaudio.Terminate()

	log.Println("Creating audio stream")
	in := make([]int16, samplesPerFrame)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	if err != nil {
		log.Fatalf("Error creating the stream: %v", err)
	}

	defer stream.Close()

	log.Println("Starting audio stream")
	err = stream.Start()
	if err != nil {
		log.Fatalf("Error starting from the stream: %v", err)
	}

	for {
		err = stream.Read()
		if err != nil {
			log.Fatalf("Error reading from the stream: %v", err)
		}
		r.Write(in)

		select {
		case <-quit:
			log.Println("Stopping audio stream")
			err := stream.Stop()
			if err != nil {
				log.Fatalf("Error stopping the stream: %v", err)
			}
			return nil
		case <-record:
			//Calling record in a separate goroutine so that the PA buffer doesn't get overflown
			go saveRecording(r.Sound())
		default:
		}
	}
}

func saveRecording(data []int16) {
	file, err := os.Create("raw_wave")
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		return
	}

	for _, sample := range data {
		//fmt.Printf("%v ", sample)
		err = binary.Write(file, binary.LittleEndian, sample)
		if err != nil {
			log.Printf("error writing to file: %v\n", err)
			return
		}
	}
	file.Close()
	log.Println("Recording saved")
}
