package main

const (
	sampleRate    = 44100
	chunkPower    = 13
	fftUpdateRate = 100

	targetRefreshRate = 120
	fftSmoothCurve    = 0.75

	minHz  = 36
	maxHz  = 20000
	minVal = 110
	maxVal = 155

	//TODO get rid of the hardcoded screen data width
	dataWidth = 64
)
