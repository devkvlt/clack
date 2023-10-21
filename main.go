package main

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	hook "github.com/robotn/gohook"
)

// sounds maps key codes to their corresponding sound names (sound files names
// without the .wav extension.)
var sounds = map[uint16]string{
	50: "a",         // `
	18: "q",         // 1
	19: "w",         // 2
	20: "e",         // 3
	21: "r",         // 4
	23: "t",         // 5
	22: "y",         // 6
	26: "u",         // 7
	28: "i",         // 8
	25: "o",         // 9
	29: "p",         // 0
	27: "t",         // -
	24: "y",         // =
	51: "backspace", // backspace

	48: "caps lock", // tab
	12: "q",         // q
	13: "w",         // w
	14: "e",         // e
	15: "r",         // r
	17: "t",         // t
	16: "y",         // y
	32: "u",         // u
	34: "i",         // i
	31: "o",         // o
	35: "p",         // p
	33: "o",         // o
	30: "p",         // p
	42: "q",         // \

	57: "caps lock", // caps lock
	0:  "a",         // a
	1:  "s",         // s
	2:  "d",         // d
	3:  "f",         // f
	5:  "g",         // g
	4:  "h",         // h
	38: "j",         // j
	40: "k",         // k
	37: "l",         // l
	41: "k",         // ;
	39: "l",         // '
	36: "enter",     // enter

	56: "enter", // shift
	6:  "z",     // z
	7:  "x",     // x
	8:  "c",     // c
	9:  "v",     // v
	11: "b",     // b
	45: "n",     // n
	46: "m",     // m
	43: "b",     // ,
	47: "n",     // .
	44: "m",     // /
	60: "enter", // rshift

	179: "z",     // fn
	59:  "x",     // ctrl
	58:  "c",     // alt
	55:  "v",     // cmd
	49:  "space", // space
	54:  "v",     // rcmd
	61:  "c",     // ralt

	123: "h", // left
	124: "l", // right
	125: "j", // down
	126: "k", // up

	53:  "a", // esc
	145: "s", // f1
	144: "d", // f2
	160: "f", // f3
	177: "g", // f4
	176: "h", // f5
	178: "j", // f6
}

// soundBuffers maps sound names to their corresponding sound buffers.
var soundBuffers = map[string]*beep.Buffer{}

//go:embed sound/*.wav
var fs embed.FS

// initSounds calls speaker.Init and populates the soundBuffers with the
// available sounds.
func initSounds() {
	var sampleRate beep.SampleRate = 44100
	bufferSize := sampleRate.N(time.Second / 10)
	speaker.Init(sampleRate, bufferSize)

	soundfiles, err := fs.ReadDir("sound")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, sf := range soundfiles {
		f, err := fs.Open("sound/" + sf.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		streamer, format, err := wav.Decode(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		soundName := strings.TrimSuffix(sf.Name(), ".wav")
		soundBuffers[soundName] = beep.NewBuffer(format)
		soundBuffers[soundName].Append(streamer)
		streamer.Close()
	}
}

// Keyboard events
const (
	KeyHold = 4
	KeyUp   = 5
)

// currentEvent keeps track of the current event for a given key code.
var currentEvent = map[uint16]uint8{}

// keyHoldHandler
func keyHoldHandler(e hook.Event) {
	if currentEvent[e.Rawcode] != KeyHold {
		currentEvent[e.Rawcode] = KeyHold
		buffer, ok := soundBuffers[sounds[e.Rawcode]]
		if ok {
			sound := buffer.Streamer(0, buffer.Len())
			speaker.Play(sound)
		} else {
			fmt.Fprintf(os.Stderr, "Error: no sound registered for the key with code %d.\n", e.Rawcode)
			os.Exit(1)
		}
	}
}

// keyUpHandler
func keyUpHandler(e hook.Event) {
	buffer := soundBuffers["release"]
	sound := buffer.Streamer(0, buffer.Len())
	speaker.Play(sound)
	currentEvent[e.Rawcode] = KeyUp
}

func main() {
	initSounds()

	hook.Register(hook.KeyHold, []string{}, keyHoldHandler)
	hook.Register(hook.KeyUp, []string{}, keyUpHandler)

	s := hook.Start()
	<-hook.Process(s)
}
