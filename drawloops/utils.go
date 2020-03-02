package drawloops

import (
	"image/color"
	"math"
)

//Linspace function returns a slice of n values which are linearly spread out between a and b.
func linspace(a, b float64, n int) []float64 {

	//At least two points are required
	if n < 2 {
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

func calculateDistance(width, height int, centerX, centerY float64) [][]float64 {
	var xx, yy, dist float64
	out := make([][]float64, width)
	for x := 0; x < width; x++ {
		out[x] = make([]float64, height)
		for y := 0; y < height; y++ {
			xx = math.Abs(float64(x) - centerX)
			yy = math.Abs(float64(y) - centerX)
			dist = math.Round(math.Hypot(xx, yy))
			out[x][y] = dist
		}
	}
	return out
}

func colorGradient(start, end color.RGBA, steps int) []color.RGBA {
	rLinspace := linspace(float64(start.R), float64(end.R), steps)
	gLinspace := linspace(float64(start.G), float64(end.G), steps)
	bLinspace := linspace(float64(start.B), float64(end.B), steps)
	aLinspace := linspace(float64(start.A), float64(end.A), steps)

	out := make([]color.RGBA, steps)
	for i := range out {
		out[i] = color.RGBA{
			R: uint8(rLinspace[i]),
			G: uint8(gLinspace[i]),
			B: uint8(bLinspace[i]),
			A: uint8(aLinspace[i]),
		}
	}

	//reverse the order of the elements in the slice
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}

	return out
}
