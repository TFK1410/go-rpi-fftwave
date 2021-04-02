package main

import (
	"fmt"
	"os"

	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
	"gopkg.in/yaml.v3"
)

type matrixConfig struct {
	Rows                   int                `yaml:"rows,omitempty"`
	Cols                   int                `yaml:"cols,omitempty"`
	Chain                  int                `yaml:"chain,omitempty"`
	Parallel               int                `yaml:"parallel,omitempty"`
	PWMBits                int                `yaml:"pwmBits,omitempty"`
	PWMLSBNanoseconds      int                `yaml:"pwmLSBNanoseconds,omitempty"`
	InitialBrightness      int                `yaml:"initialBrightness,omitempty"`
	ScanMode               rgbmatrix.ScanMode `yaml:"scanMode,omitempty"`
	DisableHardwarePulsing bool               `yaml:"disableHardwarePulsing,omitempty"`
	ShowRefresh            bool               `yaml:"showRefresh,omitempty"`
	InverseColors          bool               `yaml:"inverseColors,omitempty"`
	HardwareMapping        string             `yaml:"hardwareMapping,omitempty"`
	PixelMapperConfig      string             `yaml:"pixelMapperConfig,omitempty"`
}

type fftConfig struct {
	ChunkPower    int `yaml:"chunkPower,omitempty"`
	FFTUpdateRate int `yaml:"fftUpdateRate,omitempty"`
	BinCount      int `yaml:"binCount,omitempty"`
}

type displayConfig struct {
	RefreshRate    int     `yaml:"refreshRate,omitempty"`
	FFTSmoothCurve float64 `yaml:"fftSmoothCurve,omitempty"`
	MinHz          float64 `yaml:"minHz,omitempty"`
	MaxHz          float64 `yaml:"maxHz,omitempty"`
	MinVal         float64 `yaml:"minVal,omitempty"`
	MaxVal         float64 `yaml:"maxVal,omitempty"`
}

type whiteDotConfig struct {
	HangTime  float64 `yaml:"hangTime,omitempty"`
	DropSpeed float64 `yaml:"dropSpeed,omitempty"`
}

type soundEnergyConfig struct {
	HistoryCount int     `yaml:"historyCount,omitempty"`
	Min          float64 `yaml:"min,omitempty"`
	Max          float64 `yaml:"maxd,omitempty"`
	Saturation   int     `yaml:"saturation,omitempty"`
	HueTime      float64 `yaml:"hueTime,omitempty"`
}

type encoderConfig struct {
	DTPin         int     `yaml:"dtPin,omitempty"`
	CLKPin        int     `yaml:"clkPin,omitempty"`
	SWPin         int     `yaml:"swPin,omitempty"`
	LongPressTime float64 `yaml:"longPressTime,omitempty"`
}

type dmxConfig struct {
	SlaveAddress byte `yaml:"slaveAddress,omitempty"`
}

type lyricsOverlayConfig struct {
	RefreshRate int    `yaml:"refreshRate,omitempty"`
	LyricsDir   string `yaml:"lyricsDir,omitempty"`
	FontFile    string `yaml:"fontFile,omitempty"`
}

// Configuration is a struct holding the config of the application
// details regarding these fields can be found in config.yml
type Configuration struct {
	Matrix      *rgbmatrix.HardwareConfig
	IntMatrix   matrixConfig        `yaml:"matrixConfig"`
	SampleRate  int                 `yaml:"sampleRate"`
	FFT         fftConfig           `yaml:"fftConfig"`
	Display     displayConfig       `yaml:"displayConfig"`
	WhiteDot    whiteDotConfig      `yaml:"whiteDotConfig"`
	SoundEnergy soundEnergyConfig   `yaml:"soundEnergyConfig"`
	Encoder     encoderConfig       `yaml:"encoderConfig"`
	DMX         dmxConfig           `yaml:"dmxConfig"`
	Lyrics      lyricsOverlayConfig `yaml:"lyricsOverlayConfig"`
}

// This variable holds the default values
var cfg Configuration = Configuration{
	IntMatrix: matrixConfig{
		Rows:                   32,
		Cols:                   64,
		Chain:                  4,
		Parallel:               1,
		PWMBits:                11,
		PWMLSBNanoseconds:      130,
		InitialBrightness:      30,
		ScanMode:               rgbmatrix.Progressive,
		DisableHardwarePulsing: false,
		ShowRefresh:            false,
		InverseColors:          false,
		HardwareMapping:        "regular",
	},
	SampleRate: 44100,
	FFT: fftConfig{
		ChunkPower:    13,
		FFTUpdateRate: 100,
		BinCount:      64,
	},
	Display: displayConfig{
		RefreshRate:    120,
		FFTSmoothCurve: 0.75,
		MinHz:          36,
		MaxHz:          20000,
		MinVal:         110,
		MaxVal:         155,
	},
	WhiteDot: whiteDotConfig{
		HangTime:  0.5,
		DropSpeed: 25,
	},
	SoundEnergy: soundEnergyConfig{
		HistoryCount: 128,
		Min:          900,
		Max:          1750,
		Saturation:   100,
		HueTime:      10,
	},
	Encoder: encoderConfig{
		DTPin:         16,
		CLKPin:        20,
		SWPin:         12,
		LongPressTime: 2,
	},
	DMX: dmxConfig{
		SlaveAddress: 0x04,
	},
	Lyrics: lyricsOverlayConfig{
		RefreshRate: 30,
		LyricsDir:   "/home/pi/share/go-rpi-fftwave/lyrics",
		FontFile:    "/home/pi/share/go-rpi-fftwave/lyrics/RobotoMono-Light.ttf",
	},
}

func loadConfig(cfg *Configuration, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening the config file: %v", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error parsing the config file: %v", err)
	}

	createMatrixConfig(cfg)

	return nil
}

// createMatrixConfig converts the internal matrix config (with the yaml mappings) to the rgbmatrix.HardwareConfig format
func createMatrixConfig(cfg *Configuration) {
	cfg.Matrix = &rgbmatrix.DefaultConfig
	cfg.Matrix.Rows = cfg.IntMatrix.Rows
	cfg.Matrix.Cols = cfg.IntMatrix.Cols
	cfg.Matrix.ChainLength = cfg.IntMatrix.Chain
	cfg.Matrix.Parallel = cfg.IntMatrix.Parallel
	cfg.Matrix.PWMBits = cfg.IntMatrix.PWMBits
	cfg.Matrix.PWMLSBNanoseconds = cfg.IntMatrix.PWMLSBNanoseconds
	cfg.Matrix.Brightness = cfg.IntMatrix.InitialBrightness
	cfg.Matrix.ScanMode = cfg.IntMatrix.ScanMode
	cfg.Matrix.DisableHardwarePulsing = cfg.IntMatrix.DisableHardwarePulsing
	cfg.Matrix.ShowRefreshRate = cfg.IntMatrix.ShowRefresh
	cfg.Matrix.InverseColors = cfg.IntMatrix.InverseColors
	cfg.Matrix.HardwareMapping = cfg.IntMatrix.HardwareMapping
	cfg.Matrix.PixelMapperConfig = cfg.IntMatrix.PixelMapperConfig
}
