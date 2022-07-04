package main

import (
	"encoding/json"
	"image/color"
	"math/rand"
	"strings"

	"github.com/golang/freetype"
	"golang.org/x/image/math/fixed"
)

//DefaultFontSize Font size
const DefaultFontSize = 12

// Lyric ...
type Lyric struct {
	ID   int
	Text []string
	// if this is 2 the partial lyric display will even cut letters
	// if this is 1 then the cut will be by letter
	// if this is 0 by default the cut will be by word
	SmoothPartial byte       `json:"smoothpartial,omitempty"`
	Color         color.RGBA `json:"color,omitempty"`
	// if 2 text is aligned to the right
	// if 1 text is aligned to the left
	// if 0 by default text is aligned to center
	Align byte `json:"align,omitempty"`
	// if AlignVisible is set to true then only the visible parts of the text will be aligned
	// otherwise by default all the text will be aligned as if the whole was already there
	AlignVisible bool `json:"alignvisible,omitempty"`
	// if RandomPosition is set to true then the text will be positioned randomly on the screen
	// the initial position will be randomized on lyric load
	RandomPosition bool `json:"randomposition,omitempty"`
	LinePositions  []fixed.Point26_6
	XRandomOffset  fixed.Int26_6
	YRandomOffset  fixed.Int26_6
	// this defines the amount of text glitching
	// https://github.com/TotallyNotChase/glitch-this
	// 0 by default no glitching and level goes up to 255
	Glitch float64 `json:"glitch,omitempty"`
	// boolean flag that defines if the glitch should also offset the color
	GlitchColor bool `json:"glitchcolor,omitempty"`
	// font size - 12 by default
	Size byte `json:"size,omitempty"`
	// if this is not empty then below the primary text a secondary text will be displayed
	// start and end parameters decide when the full text will be displayed
	// size will be the font size of the secondary text
	Alt struct {
		Text  string `json:"text"`
		Start byte   `json:"start,omitempty"`
		End   byte   `json:"end,omitempty"`
		Size  byte   `json:"size,omitempty"`
	} `json:"alt,omitempty"`
}

// NewLyric default constructor
func NewLyric() *Lyric {
	l := new(Lyric)
	l.ID = -1
	l.SmoothPartial = 0
	l.Color = color.RGBA{255, 255, 255, 255}
	l.Align = 0
	l.AlignVisible = false
	l.RandomPosition = false
	l.Glitch = 0
	l.Size = DefaultFontSize
	return l
}

func (l *Lyric) getPartialText(progress byte) ([]string, int) {
	if progress == 255 {
		return l.Text, 100
	} else if progress == 0 {
		return []string{}, 100
	}

	var charCount int
	for _, s := range l.Text {
		charCount += len(s)
	}

	cutFloat := float64(charCount) * float64(progress) / 255.0
	cutCount := int(cutFloat)
	cutChar := 100
	if l.SmoothPartial == 2 {
		cutChar = int((cutFloat - float64(cutCount)) * 100)
	}

	var lineCount, lineRemainder int
	charCount = 0
	for _, s := range l.Text {
		lineCount++
		charCount += len(s)
		if charCount > int(cutCount) {
			lineRemainder = cutCount - charCount + len(s)
			break
		}
	}
	textOut := make([]string, lineCount)
	for i := 0; i < lineCount-1; i++ {
		textOut[i] = l.Text[i]
	}
	if lineRemainder == len(l.Text[lineCount-1]) {
		textOut[lineCount-1] = l.Text[lineCount-1]
	} else {
		if l.SmoothPartial == 0 {
			words := strings.Fields(l.Text[lineCount-1])
			var wordLengths, wordCount int
			for _, word := range words {
				wordLengths += len(word)
				if wordLengths < lineRemainder {
					wordCount++
				} else {
					break
				}
				wordLengths++
			}
			textOut[lineCount-1] = strings.Join(words[0:wordCount], " ")
		} else if l.SmoothPartial == 1 {
			textOut[lineCount-1] = l.Text[lineCount-1][0:lineRemainder]
		} else if l.SmoothPartial == 2 {
			textOut[lineCount-1] = l.Text[lineCount-1][0 : lineRemainder+1]
		}
	}

	return textOut, cutChar
}

func (l *Lyric) divideLyric(lyric string, sizeX int, xRatio float64) {
	var line string
	var lines []string
	charsPerLine := int(float64(sizeX) / (float64(l.Size) * xRatio))

	lyriclines := strings.Split(lyric, "\n")
	for _, lyricline := range lyriclines {
		words := strings.Split(lyricline, " ")
		for _, word := range words {
			line += word + " "
			if len(word) > int(charsPerLine)-1 {
				lines = append(lines, word)
				line = ""
			} else if len(line) > int(charsPerLine) {
				lines = append(lines, line[0:len(line)-len(word+" ")-1])
				line = word + " "
			}
		}
		if len(line) > 0 {
			lines = append(lines, line[0:len(line)-1])
		}
	}
	l.Text = lines
}

func (l *Lyric) initialPositions(ldc *LyricDrawContext) {
	// to be incremented by one if there's alt text
	l.LinePositions = make([]fixed.Point26_6, len(l.Text))

	textHeight := float64(len(l.Text)*(int(l.Size)+spacing) - spacing)
	pt := freetype.Pt(0, int(ldc.c.PointToFixed((float64(ldc.sizey)-textHeight)/2+float64(l.Size)-2)>>6))

	var i int
	for _, s := range l.Text {

		switch l.Align {
		case 0:
			// Text center
			textWidth := float64(l.Size) * ldc.xRatio * float64(len(s))
			whiteSpaceWidth := float64(ldc.sizex) - textWidth
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
			textWidth := float64(l.Size) * ldc.xRatio * float64(len(s))
			whiteSpaceWidth := float64(ldc.sizex) - textWidth
			if whiteSpaceWidth > 0 {
				pt.X = ldc.c.PointToFixed(whiteSpaceWidth)
			} else {
				pt.X = ldc.c.PointToFixed(0)
			}
		default:
		}

		l.LinePositions[i] = pt

		pt.Y += ldc.c.PointToFixed(float64(l.Size) + spacing)
		i++
	}

	if l.RandomPosition {
		var maxWidth float64
		for _, s := range l.Text {
			textWidth := float64(l.Size) * ldc.xRatio * float64(len(s))
			if textWidth > maxWidth {
				maxWidth = textWidth
			}
		}

		y := 0
		if int(textHeight) < ldc.sizey {
			yWhiteSpace := ldc.sizey - int(textHeight)
			y = rand.Intn(yWhiteSpace) - yWhiteSpace/2
		}

		x := 0
		if maxWidth < float64(ldc.sizex) {
			x = rand.Intn(ldc.sizex - int(maxWidth))
			if l.Align == 0 {
				x -= (ldc.sizex - int(maxWidth)) / 2
			} else if l.Align == 2 {
				x = 0 - x
			}
		}
		l.XRandomOffset = ldc.c.PointToFixed(float64(x))
		l.YRandomOffset = ldc.c.PointToFixed(float64(y))
		for i := 0; i < len(l.LinePositions); i++ {
			l.LinePositions[i].X += l.XRandomOffset
			l.LinePositions[i].Y += l.YRandomOffset
		}

	}
}

func (l *Lyric) parseLyricData(ID int, text string, parameters string, ldc *LyricDrawContext) {
	l.ID = ID
	l.divideLyric(text, ldc.sizex, ldc.xRatio)
	l.initialPositions(ldc)
	json.Unmarshal([]byte(parameters), &l)
}
