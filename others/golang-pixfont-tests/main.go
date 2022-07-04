package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"golang-pixfont-tests/pxl"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

// LyricDrawContext ...
type LyricDrawContext struct {
	sizex, sizey int
	img          *image.RGBA
	db           *sql.DB
}

const (
	spacing = 1  //Line spacing in pixels
	dpi     = 72 //Screen DPI the best constant for Freetype
)

func saveImage(rgba image.Image) {
	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

// Clear canvas.
func (ldc *LyricDrawContext) clear() {
	draw.Draw(ldc.img, ldc.img.Bounds(), image.Transparent, image.ZP, draw.Src)
}

// Using this function we calculate the ratio of character width to font size
// this is later used when for example centering the drawn string
// this is calculated by setting the font size to 100 points
// then we divide the point advance after the font draw by 100 and we get the ratio
// func (ldc *LyricDrawContext) getFontRatio() {
// 	ldc.clear()
// 	ldc.c.SetFontSize(100)

// 	var pt Position
// 	testString := "1234 567asdvxzc awetgbop 89"
// 	ptadv, err := ldc.c.DrawString(testString, pt)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	ldc.xRatio = (float64(ptadv.X>>6) + float64(ptadv.X&0x3f)/(1<<6)) / float64(len(testString)*100.0)

// 	ldc.c.SetFontSize(DefaultFontSize)

// 	ldc.clear()
// }

func initContext(sizex, sizey int) *LyricDrawContext {
	ldc := new(LyricDrawContext)

	ldc.sizex = sizex
	ldc.sizey = sizey
	ldc.img = image.NewRGBA(image.Rect(0, 0, sizex, sizey))

	return ldc
}

func (ldc *LyricDrawContext) drawRemainderCover(l *Lyric, c rune, x, y int, charRemainder int) {
	csize := l.getRuneLength(c)
	xAdjust := int(math.Ceil(float64(csize) * float64(100-charRemainder) / 100.0))
	rectAdjust := image.Rect(x+1-xAdjust, y, x+1, y-l.FontSize-1)
	draw.Draw(ldc.img, rectAdjust, image.Transparent, image.ZP, draw.Src)
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
	pt.Y = int((float64(ldc.sizey) - textHeight) / 2)

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
		whiteSpaceWidth := ldc.sizex - textWidth

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

func (ldc *LyricDrawContext) getImage() *image.RGBA {
	return ldc.img
}

func main() {
	ldc := initContext(128, 64)

	ldc.startSqlite()

	// a separate seed thread makes it so that the glitch shifts are changed only
	// by the defined ticker time instead of with every frame
	glitchSeed = time.Now().UTC().UnixNano()
	rand.Seed(glitchSeed)

	l := ldc.getLyric(1)

	prgs := byte(200)

	if l.AlignVisible {
		ldc.drawTextRealign(l, prgs)
	} else {
		ldc.drawText(l, prgs)
	}
	if l.Glitch >= 0.1 {
		glitchImage(ldc.img, l.Glitch, l.GlitchColor)
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetOutputMode(termbox.Output256)

	pxl.DisplayImage(ldc.getImage())
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Ch == 'q' {
				return
			}
			if ev.Key == termbox.KeyArrowLeft {
				prgs -= 2
			} else if ev.Key == termbox.KeyArrowRight {
				prgs += 2
			} else if ev.Key == termbox.KeyArrowUp {
				l.Glitch += 0.1
			} else if ev.Key == termbox.KeyArrowDown {
				l.Glitch -= 0.1
			} else if ev.Ch == 'c' {
				l.GlitchColor = !l.GlitchColor
			} else if ev.Key == termbox.KeySpace {
				time.Sleep(10 * time.Millisecond)
			}

			if l.AlignVisible {
				ldc.drawTextRealign(l, prgs)
			} else {
				ldc.drawText(l, prgs)
			}
			if l.Glitch >= 0.1 {
				glitchImage(ldc.img, l.Glitch, l.GlitchColor)
			}
			pxl.DisplayImage(ldc.getImage())
		case termbox.EventResize:
			pxl.DisplayImage(ldc.getImage())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
