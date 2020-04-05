package main

import (
	"encoding/binary"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TFK1410/go-rpi-fftwave/soundbuffer"

	"github.com/gordonklaus/portaudio"
)

func initRecord(r *soundbuffer.SoundBuffer, samplesPerFrame int, ss SoundSync) error {
	defer ss.wg.Done()

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
		// With the non callback stream reading method the buffer can sometimes overflow
		// We ignore that error and continue onto the next loop
		err = stream.Read()
		if err != nil {
			if err.Error() == "Input overflowed" {
				// log.Println("Recording input overflown. Continuing...")
				continue
			} else {
				log.Fatalf("Error reading from the stream: %v", err)
			}
		}
		r.Write(in)

		select {
		case <-ss.quit:
			// Wrap up the audio stream after the quit message is received
			log.Println("Stopping audio stream")
			err := stream.Stop()
			if err != nil {
				log.Fatalf("Error stopping the stream: %v", err)
			}
			return nil
		case <-record:
			// Calling record in a separate goroutine so that the PA buffer doesn't get overflown
			go saveRecording(r.Sound())
		case ss.sb <- r:
		default:
		}
	}
}

// saveRecording function saves the current data in the buffer to a raw_wave file
// this not at all in any sort of wave format
// however this can be read through for example Audacity with the raw wave import functions
func saveRecording(data []int16) {
	file, err := os.Create("raw_wave")
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		return
	}

	for _, sample := range data {
		err = binary.Write(file, binary.LittleEndian, sample)
		if err != nil {
			log.Printf("error writing to file: %v\n", err)
			return
		}
	}
	file.Close()
	log.Println("Recording saved")
}
