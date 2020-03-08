package main

import "time"

const (
	sampleRate    = 44100
	chunkPower    = 13
	fftUpdateRate = 100

	targetRefreshRate = 90
	fftSmoothCurve    = 0.75

	minHz  = 36
	maxHz  = 20000
	minVal = 110
	maxVal = 155

	//TODO get rid of the hardcoded screen data width
	dataWidth = 64

	whiteDotHangTime  = time.Duration(500 * time.Millisecond)
	whiteDotDropSpeed = 25

	soundEnergyHistoryCount = 128
	soundEnergyMin          = 900
	soundEnergyMax          = 1750
	soundEnergySaturation   = 100
	soundEnergyHueTime      = time.Duration(10 * time.Second)

	dtPin     = 16
	clkPin    = 20
	swPin     = 12
	longPress = time.Duration(2 * time.Second)
)
