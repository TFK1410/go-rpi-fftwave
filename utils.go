package main

import (
	"math"
)

//Logspace function returns a slice of n values which are logarithmically spread out between a and b.
//Values a and b are in logarithmic space already.
func logspace(a, b float64, n int) []float64 {

	//At least two points are required
	if n < 2 {
		return nil
	}

	//b value has to be bigger than a
	if b < a {
		return nil
	}

	//Step size
	c := (b - a) / float64(n-1)

	//Create and fill the slice
	out := make([]float64, n)
	for i := range out {
		out[i] = math.Pow(10, a+float64(i)*c)
	}

	//Fix last entry to be 10^b
	out[len(out)-1] = math.Pow(10, b)

	return out
}

//Linspace function returns a slice of n values which are linearly spread out between a and b.
func linspace(a, b float64, n int) []float64 {

	//At least two points are required
	if n < 2 {
		return nil
	}

	//b value has to be bigger than a
	if b < a {
		return nil
	}

	//Step size
	c := (b - a) / float64(n-1)

	//Create and fill the slice
	out := make([]float64, n)
	for i := range out {
		out[i] = a + float64(i)*c
	}

	//Fix last entry to be b
	out[len(out)-1] = b

	return out
}

//MaxFromRange function returns the highest value in a slice between the two indexes specified
func maxFromRange(start, end int, slc []float64) float64 {
	var mx float64
	for i := start; i < end; i++ {
		if mx < math.Abs(slc[i]) {
			mx = math.Abs(slc[i])
		}
	}
	return mx
}

//CalculateBands function returns a slice of frequency bands which will then be used to translate
//linear frequency space to a logarithmic one.
func calculateBands(minHz, maxHz float64, width int) []float64 {
	return logspace(math.Log10(minHz), math.Log10(maxHz), width+1)
}

//CalculateBins function translates linear range of frequency to appropriate fourier transform bin ranges
//in logarithmic space
func calculateBins(minHz, maxHz float64, width, sampleRate, chunkSize int) []int {
	freqBands := calculateBands(minHz, maxHz, width)

	fftBins := make([]int, width+1)
	for i := range fftBins {
		fftBins[i] = int(math.Round(float64(chunkSize) * freqBands[i] / float64(sampleRate)))
		if fftBins[i] < 1 {
			fftBins[i] = 1
		} else if fftBins[i] > 1<<chunkPower {
			fftBins[i] = 1 << chunkPower
		}
	}
	return fftBins
}

//calculateBarriers function returns the values from the result of the fourier transform
//that will be appropriate to a vertical bar level
func calculateBarriers(height int, minVal, maxVal float64) []float64 {
	space := linspace(minVal, maxVal, height)

	//reverse the order of the elements in the slice
	for left, right := 0, len(space)-1; left < right; left, right = left+1, right-1 {
		space[left], space[right] = space[right], space[left]
	}
	return space
}
