package drawloops

import (
	"image/color"
	"math"

	"github.com/TFK1410/go-rpi-fftwave/dmx"
	"github.com/TFK1410/go-rpi-fftwave/palette"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
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
	//for left, right := 0, len(space)-1; left < right; left, right = left+1, right-1 {
	//	intSpace[left], intSpace[right] = byte(math.Round(space[right])), byte(math.Round(space[left]))
	//}
	// round the values to a byte
	for i := 0; i < len(space); i++ {
		intSpace[i] = byte(math.Round(space[i]))
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

// // calculateDistance calculates the distance from the center in a radial pattern and maps the matrix position to the distance
// func calculateDistance(width, height int, centerX, centerY float64) [][]int {
// 	var xx, yy, dist float64
// 	out := make([][]int, width)
// 	for x := 0; x < width; x++ {
// 		out[x] = make([]int, height)
// 		for y := 0; y < height; y++ {
// 			xx = math.Abs(float64(x) - centerX)
// 			yy = math.Abs(float64(y) - centerY)
// 			dist = math.Round(math.Hypot(xx, yy))
// 			out[x][y] = int(dist)
// 		}
// 	}
// 	return out
// }

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

func commonDraw(m Wave, c *rgbmatrix.Canvas, dmxData dmx.DMXData, data, dots []float64) {
	var maxvalue, maxdot float64
	dataWidth, dataHeight := m.GetDataSize()
	minVal, maxVal := m.GetValueRange()
	paletteIndexes := m.GetPaletteIndexes()

	xprev := 0
	for x := 0; x < dataWidth; x++ {
		xnext := (x + 1) * len(data) / dataWidth
		maxvalue = data[xprev]
		maxdot = dots[xprev]
		// for xindex := xprev; xindex < xnext; xindex++ {
		// 	if maxvalue < data[xindex] {
		// 		maxvalue = data[xindex]
		// 		maxdot = dots[xindex]
		// 	}
		// }
		xprev = xnext

		barHeight := getBarHeight(maxvalue, dataHeight, minVal, maxVal)
		dotsHeight := getBarHeight(maxdot, dataHeight, minVal, maxVal)
		if dmxData.WhiteDots && barHeight > 0 && barHeight == dotsHeight {
			barHeight--
		}
		var phaseOffset int
		if dmxData.PalettePhaseOffset > 0 && dmxData.PaletteAngle > 0 {
			phaseOffset = int(dmxData.PalettePhaseOffset) + int(float64(dmxData.PaletteAngle)/255.0*float64(dataHeight)*float64(x))
		}

		for y := 0; y < barHeight; y++ {
			if dmxData.Color.A > 0 {
				// draw constant dmx color
				m.DrawPixels(c, x, y, dmxData.Color)
			} else if dmxData.ColorPalette > 0 {
				// draw dmx palette color
				m.DrawPixels(c, x, y, palette.Palettes[dmxData.ColorPalette][getPaletteOffsetWrap(int(paletteIndexes[y])+phaseOffset)])
			} else {
				// draw default palette color
				m.DrawPixels(c, x, y, palette.Palettes[0][paletteIndexes[y]])
			}
		}

		for y := barHeight; y < dataHeight; y++ {
			// blackout the rest
			m.DrawPixels(c, x, y, color.RGBA{0, 0, 0, 0})
		}

		if dotsHeight > 0 && dmxData.WhiteDots {
			// white dot draw
			m.DrawPixels(c, x, dotsHeight-1, color.RGBA{255, 255, 255, 255})
		}
	}
}
