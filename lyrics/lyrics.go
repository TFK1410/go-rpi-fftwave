package lyrics

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	//DefaultFontSize is the default size of the font
	DefaultFontSize = 12 //Font size
	spacing         = 1  //Line spacing in pixels
	dpi             = 72 //Screen DPI the best constant for Freetype
)

//Overlay ...
type Overlay struct {
	FontFile    string
	LyricsDir   string
	RefreshRate int
	SizeX       int
	SizeY       int
}

func loadFont(path string) (*truetype.Font, error) {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return f, nil
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

func initContext(rgba *image.RGBA, f *truetype.Font) *freetype.Context {
	// Initialize the context.
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(DefaultFontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.White)
	c.SetHinting(font.HintingNone)

	return c
}

func (o *Overlay) drawText(rgba *image.RGBA, c *freetype.Context, lyric *LyricLineData) {
	// Clear canvas.
	draw.Draw(rgba, rgba.Bounds(), image.Transparent, image.ZP, draw.Src)
	if lyric == nil {
		return
	}

	charWidth := float64(lyric.size) / 12.0 * 7.0
	c.SetFontSize(float64(lyric.size))

	// Draw the text.
	textHeight := float64(len(lyric.text))*float64(lyric.size+spacing) - spacing
	pt := freetype.Pt(0, int(c.PointToFixed((float64(o.SizeY)-textHeight)/2+float64(lyric.size)-2)>>6))
	for _, s := range lyric.text {

		// Text centering
		textWidth := charWidth*float64(len(s)) + float64(strings.Count(s, " "))
		whiteSpaceWidth := float64(o.SizeX) - textWidth
		if whiteSpaceWidth > 0 {
			pt.X = c.PointToFixed(whiteSpaceWidth / 2)
		} else {
			pt.X = c.PointToFixed(0)
		}

		_, err := c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		pt.Y += c.PointToFixed(float64(lyric.size) + spacing)
	}
}

func drawOutline(rgba *image.RGBA) {
	bounds := rgba.Bounds()

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			if r > 0 && g > 0 && b > 0 && a > 0 {
				rgba.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 0xff})

				if x > 0 {
					_, _, _, a = rgba.At(x-1, y).RGBA()
					if a == 0 {
						rgba.Set(x-1, y, color.Black)
					}
				}
				if x < bounds.Max.X-1 {
					_, _, _, a = rgba.At(x+1, y).RGBA()
					if a == 0 {
						rgba.Set(x+1, y, color.Black)
					}
				}
				if y > 0 {
					_, _, _, a = rgba.At(x, y-1).RGBA()
					if a == 0 {
						rgba.Set(x, y-1, color.Black)
					}
				}
				if y < bounds.Max.Y-1 {
					_, _, _, a = rgba.At(x, y+1).RGBA()
					if a == 0 {
						rgba.Set(x, y+1, color.Black)
					}
				}

				if x > 0 && y > 0 {
					_, _, _, a = rgba.At(x-1, y-1).RGBA()
					if a == 0 {
						rgba.Set(x-1, y-1, color.Black)
					}
				}
				if x > 0 && y < bounds.Max.Y-1 {
					_, _, _, a = rgba.At(x-1, y+1).RGBA()
					if a == 0 {
						rgba.Set(x-1, y+1, color.Black)
					}
				}
				if x < bounds.Max.X-1 && y > 0 {
					_, _, _, a = rgba.At(x+1, y-1).RGBA()
					if a == 0 {
						rgba.Set(x+1, y-1, color.Black)
					}
				}
				if x < bounds.Max.X-1 && y < bounds.Max.Y-1 {
					_, _, _, a = rgba.At(x+1, y+1).RGBA()
					if a == 0 {
						rgba.Set(x+1, y+1, color.Black)
					}
				}
			}
		}
	}
}

func resizeImage(rgba *image.RGBA) {
	bounds := LyricsOverlay.Bounds()

	upperHalf := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X/2, bounds.Max.Y)
	draw.Draw(LyricsOverlay, upperHalf, rgba, image.ZP, draw.Src)

	lowerHalf := image.Rect(bounds.Max.X/2, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	draw.Draw(LyricsOverlay, lowerHalf, rgba, image.Pt(bounds.Min.X, bounds.Max.Y), draw.Src)
}

//LyricsOverlay ...
var LyricsOverlay *image.RGBA

//InitLyricsThread ...
func (o *Overlay) InitLyricsThread(lyricsMeasure, lyricsID <-chan int, wg *sync.WaitGroup, quit <-chan struct{}) {
	defer wg.Done()

	InitLRCRegexp()
	ld := NewLRCData()

	f, err := loadFont(o.FontFile)
	if err != nil {
		log.Fatal(err)
	}

	LyricsOverlay = image.NewRGBA(image.Rect(0, 0, o.SizeX*2, o.SizeY/2))
	rgba := image.NewRGBA(image.Rect(0, 0, o.SizeX, o.SizeY))
	c := initContext(rgba, f)

	// Create a loop ticker that will refresh the current lyric at the specified rate
	ticker := time.Tick(time.Second / time.Duration(o.RefreshRate))
	startTime := time.Now()

	var curMeasure, curID, newIdx, oldIdx int
	var curDuration time.Duration
	var lyric *LyricLineData
	oldIdx = -1

	// Wait and load the first incoming lyricsID
	select {
	case curID = <-lyricsID:
		// A new song ID is received
		// If it is equal to 0 that means that no song is selected right now
		// The goroutine will be paused until a good ID with an existing .lrc file is found
		for {
			if ok := ld.ReloadLyrics(o.LyricsDir, curID, o.SizeX); ok {
				break
			}

			select {
			case curID = <-lyricsID:
			case <-quit:
				log.Println("Stopping lyrics thread")
				return
			}
		}
	case <-quit:
		log.Println("Stopping lyrics thread")
		return
	}

	for {
		select {
		case curID = <-lyricsID:
			// A new song ID is received
			// If it is equal to 0 that means that no song is selected right now
			// The goroutine will be paused until a good ID with an existing .lrc file is found
			for {
				if ok := ld.ReloadLyrics(o.LyricsDir, curID, o.SizeX); ok {
					break
				}

				select {
				case curID = <-lyricsID:
				case <-quit:
					log.Println("Stopping lyrics thread")
					return
				}
			}
		case curMeasure = <-lyricsMeasure:
			// First value is the length of the track in seconds
			// Second is the number of expected DMX increments to occur throughout the track (MSB * 255 + Rest)
			// From every color transition in Rekordbox we take the end value - 1 because the end doesn't stick for long
			// log.Println("Time:", float64(236.2/994.0)*(float64(bytes[1]-1)*255.0+float64(bytes[2])), "seconds")
			curDuration = ld.Increments * time.Duration(curMeasure)
			startTime = time.Now()
		case <-ticker:
			// Clear the text if the time since the last timer adjust is more than thrice the increment
			// This is to stop the lyric updates if new data is missing from DMX
			if time.Since(startTime) > 3*ld.Increments {
				lyric, newIdx = nil, -1
			} else {
				lyric, newIdx = ld.GetCurrentLyric(curDuration + time.Since(startTime))
			}

			if newIdx != oldIdx {
				oldIdx = newIdx

				o.drawText(rgba, c, lyric)
				drawOutline(rgba)
				resizeImage(rgba)
				// log.Println("Current time", curDuration+time.Since(startTime))
				// log.Println("Current lyric", text)
			}
		case <-quit:
			log.Println("Stopping lyrics thread")
			return
		}
	}
}
