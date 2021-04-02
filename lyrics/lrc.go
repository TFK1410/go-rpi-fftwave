package lyrics

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// First value is the length of the track in seconds
// Second is the number of expected DMX increments to occur throughout the track (MSB * 255 + Rest)
// From every color transition in Rekordbox we take the end value - 1 because the end doesn't stick for long

var (
	paramsRegexp        = &regexp.Regexp{}
	lyricsRegexp        = &regexp.Regexp{}
	lyricsOptionsRegexp = &regexp.Regexp{}
)

//LyricLineData ...
type LyricLineData struct {
	start, end time.Duration
	text       []string
	size       int
}

//LRCData ...
type LRCData struct {
	params     map[string]string
	ID         int
	Measures   int
	Length     time.Duration
	Increments time.Duration
	lyrics     []LyricLineData
}

var ld *LRCData

//NewLRCData ...
func NewLRCData() *LRCData {
	var ld LRCData
	ld.params = make(map[string]string)
	return &ld
}

//InitLRCRegexp ...
func InitLRCRegexp() {
	paramsRegexp = regexp.MustCompile(`(?m)^\[([^:\d]*):(.*)\]$`)
	lyricsRegexp = regexp.MustCompile(`(?m)^\[(\d{2}:\d{2}\.\d{2})\]([^<\n]*)( <.*>)?$`)
	lyricsOptionsRegexp = regexp.MustCompile(`(?m)^(\w*)\((.*)\)$`)
}

func parseTime(str string) (time.Duration, error) {
	split := strings.SplitN(str, ":", 2)

	minutes, err := strconv.Atoi(split[0])
	if err != nil {
		return time.Duration(0), err
	}

	seconds, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(time.Duration(minutes)*time.Minute + time.Duration(seconds*1000)*time.Millisecond), nil
}

func (ld *LRCData) parseParams(bytes []byte) {
	matches := paramsRegexp.FindAllSubmatch(bytes, -1)
	for _, match := range matches {
		ld.params[string(match[1])] = string(match[2])
	}

	for _, key := range []string{"length", "id", "measures"} {
		if _, ok := ld.params[key]; !ok {
			log.Printf("%v key does not exist in the params map, please fix the file", key)
			return
		}
	}
	ld.Length, _ = parseTime(ld.params["length"])
	ld.ID, _ = strconv.Atoi(ld.params["id"])
	ld.Measures, _ = strconv.Atoi(ld.params["measures"])
}

func divideLyric(lyric string, charsPerLine int) []string {
	var line string
	var lines []string

	words := strings.Split(lyric, " ")
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
	return lines
}

func parseLyricOptions(lld *LyricLineData, s string) {
	splitOptions := strings.Split(s, ",")
	for _, option := range splitOptions {
		matches := lyricsOptionsRegexp.FindSubmatch([]byte(option))
		if len(matches) > 2 {
			switch string(matches[1]) {
			case "size":
				lld.size, _ = strconv.Atoi(string(matches[2]))
			}
		}
	}
}

func (ld *LRCData) parseLyrics(bytes []byte, sizeX int) {
	var lld LyricLineData
	matches := lyricsRegexp.FindAllSubmatch(bytes, -1)
	for i, match := range matches {
		lld.start, _ = parseTime(string(match[1]))
		lld.size = DefaultFontSize

		if len(match[3]) > 3 {
			parseLyricOptions(&lld, string(match[3][2:len(match[3])-1]))
		}

		charWidth := float64(lld.size) / 12.0 * 7.0
		charsPerLine := int(float64(sizeX) / charWidth)
		lld.text = divideLyric(strings.ToUpper(string(match[2])), charsPerLine)

		if i > 0 {
			ld.lyrics[len(ld.lyrics)-1].end = lld.start
		}

		ld.lyrics = append(ld.lyrics, lld)
	}
	ld.lyrics[len(ld.lyrics)-1].end = ld.Length
}

//GetCurrentLyric ...
func (ld *LRCData) GetCurrentLyric(curTime time.Duration) (*LyricLineData, int) {
	for idx, line := range ld.lyrics {
		if line.start < curTime && line.end > curTime {
			return &line, idx
		}
	}
	return nil, -1
}

//ParseLRC ...
func (ld *LRCData) parseLRC(path string, sizeX int) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	ld.parseParams(b)
	ld.parseLyrics(b, sizeX)
	ld.Increments = time.Duration(float64(ld.Length) / float64(ld.Measures))

	return nil
}

func findByID(dir string, ID int) (string, error) {
	matches, err := filepath.Glob(dir + "/" + strconv.Itoa(ID) + "-*")
	if err != nil {
		return "", err
	}

	if len(matches) != 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("No file match found for ID: %v", ID)
}

//ReloadLyrics function tries to load new lyrics if the ID from DMX has changed
func (ld *LRCData) ReloadLyrics(lyricsDir string, newID, sizeX int) bool {
	if newID == 0 {
		ld.ID = 0
		return false
	}

	//Find the LRC file starting with the ID and return the first match
	match, err := findByID(lyricsDir, newID)
	if err != nil {
		log.Println(err)
		return false
	}

	//Parse the found file
	err = ld.parseLRC(match, sizeX)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
