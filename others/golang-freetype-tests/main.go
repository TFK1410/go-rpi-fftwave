package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"golang-freetype-tests/pxl"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nsf/termbox-go"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// LyricDrawContext ...
type LyricDrawContext struct {
	sizex, sizey int
	xRatio       float64
	img          *image.RGBA
	f            *truetype.Font
	c            *freetype.Context
	db           *sql.DB
}

const (
	spacing = 1  //Line spacing in pixels
	dpi     = 72 //Screen DPI the best constant for Freetype
)

func (ldc *LyricDrawContext) loadFont(path string) error {
	// Open font file
	fontReader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fontReader.Close()
	// Read the font data.
	fontBytes, err := io.ReadAll(fontReader)
	if err != nil {
		return err
	}
	ldc.f, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}
	return nil
}

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
func (ldc *LyricDrawContext) getFontRatio() {
	ldc.clear()
	ldc.c.SetFontSize(100)

	pt := freetype.Pt(0, 0)
	testString := "1234 567asdvxzc awetgbop 89"
	ptadv, err := ldc.c.DrawString(testString, pt)
	if err != nil {
		log.Println(err)
		return
	}

	ldc.xRatio = (float64(ptadv.X>>6) + float64(ptadv.X&0x3f)/(1<<6)) / float64(len(testString)*100.0)

	ldc.c.SetFontSize(DefaultFontSize)

	ldc.clear()
}

func initContext(sizex, sizey int, fontpath string) *LyricDrawContext {
	ldc := new(LyricDrawContext)

	ldc.sizex = sizex
	ldc.sizey = sizey
	ldc.img = image.NewRGBA(image.Rect(0, 0, sizex, sizey))

	err := ldc.loadFont(fontpath)
	if err != nil {
		log.Fatal(err)
	}
	// Initialize the context.
	ldc.c = freetype.NewContext()
	ldc.c.SetDPI(dpi)
	ldc.c.SetFont(ldc.f)
	ldc.c.SetFontSize(DefaultFontSize)
	ldc.c.SetClip(ldc.img.Bounds())
	ldc.c.SetDst(ldc.img)
	ldc.c.SetSrc(image.White)
	ldc.c.SetHinting(font.HintingNone)

	ldc.getFontRatio()

	return ldc
}

func (ldc *LyricDrawContext) drawRemainderCover(l *Lyric, ptrem fixed.Point26_6, charRemainder int) {
	xAdjust := int(math.Ceil(float64(l.Size) * ldc.xRatio * float64(100-charRemainder) / 100.0))
	xCur := int(ptrem.X >> 6)
	yCur := int(ptrem.Y>>6) + int(float64(l.Size)*0.25)
	rectAdjust := image.Rect(xCur+1-xAdjust, yCur, xCur+1, yCur-int(l.Size))
	draw.Draw(ldc.img, rectAdjust, image.Transparent, image.ZP, draw.Src)
}

func (ldc *LyricDrawContext) drawText(l *Lyric, progress byte) {
	ldc.clear()

	ldc.c.SetFontSize(float64(l.Size))
	ldc.c.SetSrc(image.NewUniform(l.Color))

	textToDraw, charRemainder := l.getPartialText(progress)

	var ptrem fixed.Point26_6
	var err error
	for i, s := range textToDraw {
		ptrem, err = ldc.c.DrawString(s, l.LinePositions[i])
		if err != nil {
			log.Println(err)
			return
		}
	}

	if charRemainder < 100 {
		ldc.drawRemainderCover(l, ptrem, charRemainder)
	}
}

func (ldc *LyricDrawContext) drawTextRealign(l *Lyric, progress byte) {
	ldc.clear()

	ldc.c.SetFontSize(float64(l.Size))
	ldc.c.SetSrc(image.NewUniform(l.Color))

	textToDraw, charRemainder := l.getPartialText(progress)

	var pt fixed.Point26_6
	textHeight := float64(len(textToDraw)*(int(l.Size)+spacing) - spacing)
	pt = freetype.Pt(0, int(ldc.c.PointToFixed((float64(ldc.sizey)-textHeight)/2+float64(l.Size)-2)>>6))
	pt.Y += l.YRandomOffset

	var ptrem fixed.Point26_6
	var err error
	for i, s := range textToDraw {
		var textWidth float64
		if i == len(textToDraw)-1 {
			textWidth = float64(l.Size) * ldc.xRatio * (float64(len(s)) - 1 + float64(charRemainder)/100)
		} else {
			textWidth = float64(l.Size) * ldc.xRatio * float64(len(s))
		}
		whiteSpaceWidth := float64(ldc.sizex) - textWidth

		switch l.Align {
		case 0:
			// Text center
			if whiteSpaceWidth > 0 {
				pt.X = ldc.c.PointToFixed(whiteSpaceWidth / 2)
			} else {
				pt.X = ldc.c.PointToFixed(0)
			}
		case 1:
			// Text left
			pt.X = ldc.c.PointToFixed(0)
		case 2:
			// Text right
			if whiteSpaceWidth > 0 {
				pt.X = ldc.c.PointToFixed(whiteSpaceWidth) + l.XRandomOffset
			} else {
				pt.X = ldc.c.PointToFixed(0)
			}
		default:
		}

		ptrem, err = ldc.c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += ldc.c.PointToFixed(float64(l.Size) + spacing)
	}

	if charRemainder < 100 {
		ldc.drawRemainderCover(l, ptrem, charRemainder)
	}
}

func (ldc *LyricDrawContext) getImage() *image.RGBA {
	return ldc.img
}

func main() {
	ldc := initContext(128, 64, "RobotoMono-Light.ttf")

	ldc.startSqlite()

	// a separate seed thread makes it so that the glitch shifts are changed only
	// by the defined ticker time instead of with every frame
	glitchSeed = time.Now().UTC().UnixNano()
	rand.Seed(glitchSeed)

	// l := NewLyric()
	// l.SmoothPartial = 2
	// l.Glitch = 0
	// l.GlitchColor = true
	// l.Color = color.RGBA{255, 255, 0, 255}
	// l.RandomPosition = false
	// l.Align = 0
	// l.AlignVisible = false

	// s := "Tcieomv s aw gs asd e\n23 1 bsd q"
	// l.parseLyricData(0, s, "", ldc)

	l := ldc.getLyric(1)

	prgs := byte(255)

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
