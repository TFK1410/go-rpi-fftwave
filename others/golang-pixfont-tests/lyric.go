package main

import (
	"encoding/json"
	"image/color"
	"math/rand"
	"strings"

	"golang-pixfont-tests/fonts/pixelmix10"
	"golang-pixfont-tests/fonts/pixelmix13"
	"golang-pixfont-tests/fonts/pixelmix16"
	"golang-pixfont-tests/fonts/pixelmix8"

	"github.com/pbnjay/pixfont"
)

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
	LinePositions  []Position
	RandomOffset   Position
	// this defines the amount of text glitching
	// https://github.com/TotallyNotChase/glitch-this
	// 0 by default no glitching and level goes up to 255
	Glitch float64 `json:"glitch,omitempty"`
	// boolean flag that defines if the glitch should also offset the color
	GlitchColor bool `json:"glitchcolor,omitempty"`
	// font size:
	// small  - 8pt by default
	// medium - 10pt
	// large  - 13pt
	// huge   - 16pt
	Size     string `json:"size,omitempty"`
	FontSize int
	// if this is not empty then below the primary text a secondary text will be displayed
	// start and end parameters decide when the full text will be displayed
	// size will be the font size of the secondary text
	Alt struct {
		Text     string `json:"text"`
		Start    byte   `json:"start,omitempty"`
		End      byte   `json:"end,omitempty"`
		Size     string `json:"size,omitempty"`
		FontSize int
	} `json:"alt,omitempty"`
}

type Position struct {
	X, Y int
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
	l.Size = "small"
	l.FontSize = 8
	return l
}

func (l *Lyric) getStringLength(s string) int {
	switch l.FontSize {
	default:
		return pixelmix8.Font.MeasureString(s)
	case 10:
		return pixelmix10.Font.MeasureString(s)
	case 13:
		return pixelmix13.Font.MeasureString(s)
	case 16:
		return pixelmix16.Font.MeasureString(s)
	}
}

func (l *Lyric) getRuneLength(r rune) int {
	var sz int
	switch l.FontSize {
	default:
		_, sz = pixelmix8.Font.MeasureRune(r)
	case 10:
		_, sz = pixelmix10.Font.MeasureRune(r)
		break
	case 13:
		_, sz = pixelmix13.Font.MeasureRune(r)
		break
	case 16:
		_, sz = pixelmix16.Font.MeasureRune(r)
		break
	}
	return sz
}

func (l *Lyric) DrawString(dr pixfont.Drawable, x, y int, s string, clr color.Color) int {
	switch l.FontSize {
	default:
		return pixelmix8.Font.DrawString(dr, x, y, s, clr)
	case 10:
		return pixelmix10.Font.DrawString(dr, x, y, s, clr)
	case 13:
		return pixelmix13.Font.DrawString(dr, x, y, s, clr)
	case 16:
		return pixelmix16.Font.DrawString(dr, x, y, s, clr)
	}
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

func (l *Lyric) divideLyric(lyric string, sizeX int) {
	var line string
	var lines []string

	lyriclines := strings.Split(lyric, "\n")
	for _, lyricline := range lyriclines {
		words := strings.Split(lyricline, " ")
		for _, word := range words {
			line += word + " "

			wordsize := l.getStringLength(word)
			linesize := l.getStringLength(line)

			if wordsize > sizeX {
				lines = append(lines, word)
				line = ""
			} else if linesize > sizeX {
				lines = append(lines, line[0:len(line)-len(word)-2])
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
	l.LinePositions = make([]Position, len(l.Text))

	textHeight := float64(len(l.Text)*(l.FontSize+spacing) - spacing)
	var pt Position
	pt.Y = int((float64(ldc.sizey) - textHeight) / 2)

	var i int
	for _, s := range l.Text {

		switch l.Align {
		case 0:
			// Text center
			textWidth := l.getStringLength(s)
			whiteSpaceWidth := ldc.sizex - textWidth
			if whiteSpaceWidth > 0 {
				pt.X = int(whiteSpaceWidth / 2)
			} else {
				pt.X = 0
			}
		case 1:
			// Text left
			pt.X = 0
		case 2:
			// Text right
			textWidth := l.getStringLength(s)
			whiteSpaceWidth := ldc.sizex - textWidth
			if whiteSpaceWidth > 0 {
				pt.X = int(whiteSpaceWidth)
			} else {
				pt.X = 0
			}
		default:
		}

		l.LinePositions[i] = pt

		pt.Y += l.FontSize + spacing
		i++
	}

	if l.RandomPosition {
		var maxWidth int
		for _, s := range l.Text {
			textWidth := l.getStringLength(s)
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
		if maxWidth < ldc.sizex {
			x = rand.Intn(ldc.sizex - maxWidth)
			if l.Align == 0 {
				x -= (ldc.sizex - maxWidth) / 2
			} else if l.Align == 2 {
				x = 0 - x
			}
		}
		l.RandomOffset.X = x
		l.RandomOffset.Y = y
		for i := 0; i < len(l.LinePositions); i++ {
			l.LinePositions[i].X += l.RandomOffset.X
			l.LinePositions[i].Y += l.RandomOffset.Y
		}

	}
}

func (l *Lyric) parseLyricData(ID int, text string, parameters string, ldc *LyricDrawContext) {
	l.ID = ID
	json.Unmarshal([]byte(parameters), &l)

	if l.Size == "medium" {
		l.FontSize = 10
	} else if l.Size == "large" {
		l.FontSize = 13
	} else if l.Size == "huge" {
		l.FontSize = 16
	}

	l.divideLyric(text, ldc.sizex)
	l.initialPositions(ldc)
}
