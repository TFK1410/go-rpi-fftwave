package drawloops

import (
	"math"
)

// Linspace function returns a slice of n values which are linearly spread out between a and b.
func linspace(a, b float64, n int) []float64 {

	// At least two points are required
	if n < 2 {
		return nil
	}

	// Step size
	c := (b - a) / float64(n-1)

	// Create and fill the slice
	out := make([]float64, n)
	for i := range out {
		out[i] = a + float64(i)*c
	}

	// Fix last entry to be b
	out[len(out)-1] = b

	return out
}

func getBarHeight(val float64, heightSteps int, minVal, maxVal float64) int {
	if val < minVal {
		return 0
	}
	if val > maxVal {
		return heightSteps
	}

	return int((val - minVal) / (maxVal - minVal) * float64(heightSteps))
}

// calculateBarriers function returns the values from the result of the fourier transform
// that will be appropriate to a vertical bar level
func calculatePaletteIndexes(height int) []byte {
	space := linspace(0, 255, height)
	intSpace := make([]byte, height)

	// reverse the order of the elements in the slice
	for left, right := 0, len(space)-1; left < right; left, right = left+1, right-1 {
		intSpace[left], intSpace[right] = byte(math.Round(space[right])), byte(math.Round(space[left]))
	}
	return intSpace
}

// calculateBarriers function returns the values from the result of the fourier transform
// that will be appropriate to a vertical bar level
// func calculateBarriers(height int, minVal, maxVal float64) []float64 {
// 	space := linspace(minVal, maxVal, height)

// 	// reverse the order of the elements in the slice
// 	for left, right := 0, len(space)-1; left < right; left, right = left+1, right-1 {
// 		space[left], space[right] = space[right], space[left]
// 	}
// 	return space
// }

// calculateDistance calculates the distance from the center in a radial pattern and maps the matrix position to the distance
func calculateDistance(width, height int, centerX, centerY float64) [][]int {
	var xx, yy, dist float64
	out := make([][]int, width)
	for x := 0; x < width; x++ {
		out[x] = make([]int, height)
		for y := 0; y < height; y++ {
			xx = math.Abs(float64(x) - centerX)
			yy = math.Abs(float64(y) - centerY)
			dist = math.Round(math.Hypot(xx, yy))
			out[x][y] = int(dist)
		}
	}
	return out
}

// colorGradient creates a color gradient based on the start and end colors and the number of steps
// func colorGradient(start, end color.RGBA, steps int) []color.RGBA {
// 	rLinspace := linspace(float64(start.R), float64(end.R), steps)
// 	gLinspace := linspace(float64(start.G), float64(end.G), steps)
// 	bLinspace := linspace(float64(start.B), float64(end.B), steps)
// 	aLinspace := linspace(float64(start.A), float64(end.A), steps)

// 	out := make([]color.RGBA, steps)
// 	for i := range out {
// 		out[i] = color.RGBA{
// 			R: uint8(rLinspace[i]),
// 			G: uint8(gLinspace[i]),
// 			B: uint8(bLinspace[i]),
// 			A: uint8(aLinspace[i]),
// 		}
// 	}

// 	// reverse the order of the elements in the slice
// 	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
// 		out[left], out[right] = out[right], out[left]
// 	}

// 	return out
// }

func getPaletteOffsetWrap(offset int) byte {
	rotations := offset / 256
	byteCutoff := byte(offset)
	if rotations%2 == 0 {
		return byteCutoff
	} else {
		return 255 - byteCutoff
	}
}
