package lyricsoverlay

import (
	"database/sql"
	"image"
	"image/draw"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// LyricDrawContext ...
type LyricDrawContext struct {
	SizeX, SizeY int
	SqlitePath   string
	RefreshRate  int
	img          *image.RGBA
	db           *sql.DB
}

const spacing = 1 //Line spacing in pixels

// func saveImage(rgba image.Image) {
// 	outFile, err := os.Create("out.png")
// 	if err != nil {
// 		log.Println(err)
// 		os.Exit(1)
// 	}
// 	defer outFile.Close()
// 	b := bufio.NewWriter(outFile)
// 	err = png.Encode(b, rgba)
// 	if err != nil {
// 		log.Println(err)
// 		os.Exit(1)
// 	}
// 	err = b.Flush()
// 	if err != nil {
// 		log.Println(err)
// 		os.Exit(1)
// 	}
// 	fmt.Println("Wrote out.png OK.")
// }

// Clear canvas.
func (ldc *LyricDrawContext) clear() {
	draw.Draw(ldc.img, ldc.img.Bounds(), image.Transparent, image.Point{0, 0}, draw.Src)
}

func (ldc *LyricDrawContext) drawRemainderCover(l *Lyric, c rune, x, y int, charRemainder int) {
	csize := l.getRuneLength(c)
	xAdjust := int(math.Ceil(float64(csize) * float64(100-charRemainder) / 100.0))
	rectAdjust := image.Rect(x+1-xAdjust, y, x+1, y-l.FontSize-1)
	draw.Draw(ldc.img, rectAdjust, image.Transparent, image.Point{0, 0}, draw.Src)
}

func (ldc *LyricDrawContext) drawText(l *Lyric, progress byte) {
	ldc.clear()

	textToDraw, charRemainder := l.getPartialText(progress)

	var lastchar rune
	var x, y int
	for i, s := range textToDraw {
		x = l.LinePositions[i].X
		y = l.LinePositions[i].Y
		x = l.DrawString(ldc.img, x, y, s, l.Color)
		lastchar = rune(s[len(s)-1])
	}

	if charRemainder < 100 {
		ldc.drawRemainderCover(l, lastchar, x, y+l.FontSize+spacing, charRemainder)
	}
}

func (ldc *LyricDrawContext) drawTextRealign(l *Lyric, progress byte) {
	ldc.clear()

	textToDraw, charRemainder := l.getPartialText(progress)

	textHeight := float64(len(textToDraw)*(l.FontSize+spacing) - spacing)
	var pt Position
	pt.Y = int((float64(ldc.SizeY) - textHeight) / 2)

	var err error
	var lastchar rune
	for i, s := range textToDraw {
		var textWidth int
		if i == len(textToDraw)-1 {
			if len(s) > 0 {
				lastchar := l.getRuneLength(rune(s[len(s)-1]))
				textWidth = l.getStringLength(s) - int(float64(lastchar)*float64(100-charRemainder)/100)
			}
		} else {
			textWidth = l.getStringLength(s)
		}
		whiteSpaceWidth := ldc.SizeX - textWidth

		switch l.Align {
		case 0:
			// Text center
			if whiteSpaceWidth > 0 {
				pt.X = whiteSpaceWidth / 2
			} else {
				pt.X = 0
			}
		case 1:
			// Text left
			pt.X = 0
		case 2:
			// Text right
			if whiteSpaceWidth > 0 {
				pt.X = whiteSpaceWidth + l.RandomOffset.X
			} else {
				pt.X = 0
			}
		default:
		}

		pt.X = l.DrawString(ldc.img, pt.X, pt.Y, s, l.Color)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += l.FontSize + spacing

		if len(s) > 0 {
			lastchar = rune(s[len(s)-1])
		}
	}

	if charRemainder < 100 {
		ldc.drawRemainderCover(l, lastchar, pt.X, pt.Y, charRemainder)
	}
}

// func (ldc *LyricDrawContext) drawOutline() {
// 	bounds := ldc.img.Bounds()

// 	for x := 0; x < bounds.Max.X; x++ {
// 		for y := 0; y < bounds.Max.Y; y++ {
// 			r, g, b, a := ldc.img.At(x, y).RGBA()
// 			if r > 0 && g > 0 && b > 0 && a > 0 {
// 				ldc.img.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 0xff})

// 				if x > 0 {
// 					_, _, _, a = ldc.img.At(x-1, y).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x-1, y, color.Black)
// 					}
// 				}
// 				if x < bounds.Max.X-1 {
// 					_, _, _, a = ldc.img.At(x+1, y).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x+1, y, color.Black)
// 					}
// 				}
// 				if y > 0 {
// 					_, _, _, a = ldc.img.At(x, y-1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x, y-1, color.Black)
// 					}
// 				}
// 				if y < bounds.Max.Y-1 {
// 					_, _, _, a = ldc.img.At(x, y+1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x, y+1, color.Black)
// 					}
// 				}

// 				if x > 0 && y > 0 {
// 					_, _, _, a = ldc.img.At(x-1, y-1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x-1, y-1, color.Black)
// 					}
// 				}
// 				if x > 0 && y < bounds.Max.Y-1 {
// 					_, _, _, a = ldc.img.At(x-1, y+1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x-1, y+1, color.Black)
// 					}
// 				}
// 				if x < bounds.Max.X-1 && y > 0 {
// 					_, _, _, a = ldc.img.At(x+1, y-1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x+1, y-1, color.Black)
// 					}
// 				}
// 				if x < bounds.Max.X-1 && y < bounds.Max.Y-1 {
// 					_, _, _, a = ldc.img.At(x+1, y+1).RGBA()
// 					if a == 0 {
// 						ldc.img.Set(x+1, y+1, color.Black)
// 					}
// 				}
// 			}
// 		}
// 	}
// }

func (ldc *LyricDrawContext) GetImage() *image.RGBA {
	return ldc.img
}

func (ldc *LyricDrawContext) InitLyricsThread(lyricProgress <-chan byte, lyricsID <-chan int, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()
	ldc.img = image.NewRGBA(image.Rect(0, 0, ldc.SizeX, ldc.SizeY))

	ldc.startSqlite()

	// a separate seed thread makes it so that the glitch shifts are changed only
	// by the defined ticker time instead of with every frame
	glitchSeed = time.Now().UTC().UnixNano()
	rand.Seed(glitchSeed)

	// Create a loop ticker that will refresh the current lyric at the specified rate
	ticker := time.NewTicker(time.Second / time.Duration(ldc.RefreshRate))
	// startTime := time.Now()

	var curID, oldID int
	// var curDuration time.Duration
	var l *Lyric
	var curProgress, oldProgress byte

	for {
		select {
		case curID = <-lyricsID:
			// A new song ID is received
			// If it is equal to 0 that means that no song is selected right now
			// The goroutine will be paused until a good ID with an existing .lrc file is found
			for {
				if l = ldc.getLyric(curID); l != nil {
					curProgress = 0
					break
				}

				select {
				case curID = <-lyricsID:
				case <-quit:
					log.Println("Stopping lyrics thread")
					return
				}
			}
		case curProgress = <-lyricProgress:
		case <-ticker.C:
			// Clear the text if the time since the last timer adjust is more than thrice the increment
			// This is to stop the lyric updates if new data is missing from DMX
			// if time.Since(startTime) > 3*ld.Increments {
			// 	lyric, newIdx = nil, -1
			// } else {
			// 	lyric, newIdx = ld.GetCurrentLyric(curDuration + time.Since(startTime))
			// }

			if curID > 0 && curProgress > 0 {
				if oldID != curID || oldProgress != curProgress {
					if l.AlignVisible {
						ldc.drawTextRealign(l, curProgress)
					} else {
						ldc.drawText(l, curProgress)
					}
					if l.Glitch >= 0.1 {
						glitchImage(ldc.img, l.Glitch, l.GlitchColor)
					}

					oldID = curID
					oldProgress = curProgress
				}
			} else {
				ldc.clear()
			}
		case <-quit:
			log.Println("Stopping lyrics thread")
			return
		}
	}
}
