// HEAVILY PORTED FROM THE PYTHON LIBRARY https://github.com/TotallyNotChase/glitch-this

package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

var glitchSeed int64
var glitchSeedTimer time.Time

// glitchImage Glitches the image
// Intensity of glitch depends on glitch_amount
func glitchImage(img *image.RGBA, glitchAmount float64, colorOffsetFlag bool) {
	// glitchAmount [0.1,10] constrain
	if glitchAmount < 0.1 {
		glitchAmount = 0.1
	}
	if glitchAmount > 10 {
		glitchAmount = 10
	}

	// set the seed
	if time.Now().Sub(glitchSeedTimer) > 500*time.Millisecond {
		glitchSeedTimer = time.Now()
		glitchSeed = time.Now().UTC().UnixNano()
	}
	rand.Seed(glitchSeed)

	maxOffset := int(math.Pow(glitchAmount, 2) / 100 * float64(img.Bounds().Size().X))
	doubledGlitchAmount := glitchAmount * 2
	imageTensor := imageToTensor(img)
	for shiftNumber := 0; shiftNumber < int(doubledGlitchAmount); shiftNumber++ {

		// Setting up offset needed for the randomized glitching
		currentOffset := rand.Intn(2*maxOffset+1) - maxOffset

		if currentOffset == 0 {
			// Can't wrap left OR right when offset is 0, End of Array
			continue
		}

		if currentOffset < 0 {
			// Grab a rectangle of specific width and heigh, shift it left
			// by a specified offset
			// Wrap around the lost pixel data from the right
			glitchLeft(imageTensor, -currentOffset)
		} else {
			// Grab a rectangle of specific width and height, shift it right
			// by a specified offset
			// Wrap around the lost pixel data from the left
			glitchRight(imageTensor, currentOffset)
		}
	}

	if colorOffsetFlag {
		// Get the next random channel we'll offset, needs to be before the random.randints
		// arguments because they will use up the original seed (if a custom seed is used)
		// the channels of choice are 'R', 'G', 'B' with 'R' being 0th channel
		randomChannel := rand.Intn(3)
		// Add color channel offset if checked true
		colorOffset(imageTensor, rand.Intn(int(2*doubledGlitchAmount)+1)-int(doubledGlitchAmount),
			rand.Intn(int(2*doubledGlitchAmount)+1)-int(doubledGlitchAmount), randomChannel)
	}

	tensorToImage(img, imageTensor)
}

func imageToTensor(img *image.RGBA) [][]color.RGBA {
	pxl := img.Pix
	dx := img.Rect.Dx()
	dy := img.Rect.Dy()
	stride := img.Stride
	a := make([][]color.RGBA, dy)
	for y := range a {
		a[y] = make([]color.RGBA, dx)
		for x := range a[y] {
			index := y*stride + x*4
			a[y][x] = color.RGBA{pxl[index], pxl[index+1], pxl[index+2], pxl[index+3]}
		}
	}

	return a
}

func tensorToImage(img *image.RGBA, tensor [][]color.RGBA) {
	pxl := img.Pix
	stride := img.Stride
	for y := range tensor {
		for x := range tensor[0] {
			index := y*stride + x*4
			pxl[index] = tensor[y][x].R
			pxl[index+1] = tensor[y][x].G
			pxl[index+2] = tensor[y][x].B
			pxl[index+3] = tensor[y][x].A
		}
	}
	img.Pix = pxl
}

func getChunk(img [][]color.RGBA, startY, stopY, startX, stopX int) [][]color.RGBA {
	chunk := make([][]color.RGBA, stopY-startY)
	for i := range chunk {
		chunk[i] = make([]color.RGBA, stopX-startX)
		copy(chunk[i], img[startY+i][startX:stopX])
	}
	return chunk
}

func setChunk(img, chunk [][]color.RGBA, startY, stopY, startX, stopX int) {
	for i := 0; i < stopY-startY; i++ {
		copy(img[startY+i][startX:stopX], chunk[i])
	}
}

// glitchLeft Grabs a rectangle from img and shifts it leftwards
// Any lost pixel data is wrapped back to the right
// Rectangle's Width and Height are determined from offset
func glitchLeft(img [][]color.RGBA, offset int) {
	// Setting up values that will determine the rectangle height
	sizeY := len(img)
	sizeX := len(img[0])
	startY := rand.Intn(sizeY + 1)
	chunkHeight := rand.Intn(int(sizeY/4)) + 1
	chunkHeight = min(chunkHeight, sizeY-startY)
	stopY := startY + chunkHeight

	// For copy
	startX := offset
	// For paste
	stopX := sizeX - startX

	leftChunk := getChunk(img, startY, stopY, startX, sizeX)
	wrapChunk := getChunk(img, startY, stopY, 0, startX)
	setChunk(img, leftChunk, startY, stopY, 0, stopX)
	setChunk(img, wrapChunk, startY, stopY, stopX, sizeX)
}

// glitchRight Grabs a rectangle from img and shifts it rightwards
// Any lost pixel data is wrapped back to the left
// Rectangle's Width and Height are determined from offset
func glitchRight(img [][]color.RGBA, offset int) {
	// Setting up values that will determine the rectangle height
	sizeY := len(img)
	sizeX := len(img[0])
	startY := rand.Intn(sizeY)
	chunkHeight := rand.Intn(int(sizeY/4)-1) + 1
	chunkHeight = min(chunkHeight, sizeY-startY)
	stopY := startY + chunkHeight

	// For copy
	stopX := sizeX - offset
	// For paste
	startX := offset

	rightChunk := getChunk(img, startY, stopY, 0, stopX)
	wrapChunk := getChunk(img, startY, stopY, stopX, sizeX)
	setChunk(img, rightChunk, startY, stopY, startX, sizeX)
	setChunk(img, wrapChunk, startY, stopY, 0, startX)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Takes the given channel's color value from imageTensor,
// starting from (0, 0)
// and puts it in the same channel's slot back in imageTensor,
// starting from (offsetY, offsetX)
func colorOffset(img [][]color.RGBA, offsetX, offsetY, channelIndex int) {
	sizeY := len(img)
	sizeX := len(img[0])
	// Make sure offsetX isn't negative in the actual algo
	if offsetX < 0 {
		offsetX += sizeX
	}
	if offsetY < 0 {
		offsetY += sizeY
	}

	imgCopy := make([][]color.RGBA, sizeY)
	for i := range imgCopy {
		imgCopy[i] = make([]color.RGBA, sizeX)
		copy(imgCopy[i], img[i])
	}

	// Assign values from 0th row of inputarr to offset_y th
	// row of outputarr
	// If outputarr's columns run out before inputarr's does,
	// wrap the remaining values around

	// Continue afterwards till end of outputarr
	// Make sure the width and height match for both slices

	// Restart from 0th row of outputarr and go until the offset_y th row
	// This will assign the remaining values in inputarr to outputarr
	if channelIndex == 0 {
		for i := 0; i < sizeX-offsetX; i++ {
			img[offsetY][offsetX+i].R = imgCopy[0][i].R
		}
		for i := 0; i < offsetX; i++ {
			img[offsetY][i].R = imgCopy[0][sizeX-offsetX+i].R
		}
		for i := 0; i < sizeY-offsetY-1; i++ {
			for j := 0; j < sizeX; j++ {
				img[offsetY+1+i][j].R = imgCopy[i+1][j].R
			}
		}
		for i := 0; i < offsetY; i++ {
			for j := 0; j < sizeX; j++ {
				img[i][j].R = imgCopy[sizeY-offsetY+i][j].R
			}
		}
	} else if channelIndex == 1 {
		for i := 0; i < sizeX-offsetX; i++ {
			img[offsetY][offsetX+i].G = imgCopy[0][i].G
		}
		for i := 0; i < offsetX; i++ {
			img[offsetY][i].G = imgCopy[0][sizeX-offsetX+i].G
		}
		for i := 0; i < sizeY-offsetY-1; i++ {
			for j := 0; j < sizeX; j++ {
				img[offsetY+1+i][j].G = imgCopy[i+1][j].G
			}
		}
		for i := 0; i < offsetY; i++ {
			for j := 0; j < sizeX; j++ {
				img[i][j].G = imgCopy[sizeY-offsetY+i][j].G
			}
		}
	} else if channelIndex == 2 {
		for i := 0; i < sizeX-offsetX; i++ {
			img[offsetY][offsetX+i].B = imgCopy[0][i].B
		}
		for i := 0; i < offsetX; i++ {
			img[offsetY][i].B = imgCopy[0][sizeX-offsetX+i].B
		}
		for i := 0; i < sizeY-offsetY-1; i++ {
			for j := 0; j < sizeX; j++ {
				img[offsetY+1+i][j].B = imgCopy[i+1][j].B
			}
		}
		for i := 0; i < offsetY; i++ {
			for j := 0; j < sizeX; j++ {
				img[i][j].B = imgCopy[sizeY-offsetY+i][j].B
			}
		}
	}
}
