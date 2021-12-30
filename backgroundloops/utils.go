package backgroundloops

import (
	"image/color"
	"math"
)

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

// maxTriBand returns the max value from the tri band struct
func maxTriBand(b SoundEnergyTriBand) float64 {
	mx := b.Bass
	if mx < b.Mid {
		mx = b.Mid
	}
	if mx < b.Treble {
		mx = b.Treble
	}
	return mx
}

// // Based on time the sound energy values are being translated from HSV to RGB values
// func soundHue(rotationTime time.Duration, soundEnergy, min, max float64) color.RGBA {
// 	var H, S, V float64

// 	H = float64(time.Now().UnixMilli()) / float64(rotationTime.Milliseconds())
// 	S = 0.7
// 	V = (soundEnergy - min) / (max - min)

// 	if V < 0 {
// 		V = 0
// 	} else if V > 1 {
// 		V = 1
// 	}

// 	return hsv2RGB(H, S, V)

// }

// H S V parameters are all between 0 and 1
func hsv2RGB(H, S, V float64) color.RGBA {
	var r, g, b float64
	if S == 0 {
		r = V * 255
		g = V * 255
		b = V * 255
	} else {
		h := H * 6
		if h == 6 {
			h = 0
		}
		i := math.Floor(h)
		v1 := V * (1 - S)
		v2 := V * (1 - S*(h-i))
		v3 := V * (1 - S*(1-(h-i)))

		if i == 0 {
			r = V
			g = v3
			b = v1
		} else if i == 1 {
			r = v2
			g = V
			b = v1
		} else if i == 2 {
			r = v1
			g = V
			b = v3
		} else if i == 3 {
			r = v1
			g = v2
			b = V
		} else if i == 4 {
			r = v3
			g = v1
			b = V
		} else {
			r = V
			g = v1
			b = v2
		}
		// RGB results from 0 to 255
		r = r * 255
		g = g * 255
		b = b * 255
	}
	out := color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
	return out
}
