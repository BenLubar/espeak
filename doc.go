//go:generate build/make_js.bash

package espeak // import "gopkg.in/BenLubar/espeak.v2"

import (
	"sync"
	"time"
)

type Error struct {
	Code    uint32
	Message string
}

func (err *Error) Error() string {
	return "espeak: " + err.Message
}

func SampleRate() int {
	return getSampleRate()
}

var lock sync.Mutex

type Context struct {
	Samples []int16
	Events  []*SynthEvent

	rate   int // words per minute, 80 to 450; default 175
	volume int // percentage of normal volume, min 0; default 100
	pitch  int // base pitch, 0 to 100; default 50
	tone   int // pitch range, 0 to 100; 0 is monotone; default 50
	// punctuation?
	// capitals?
	// word gap?

	voice struct {
		name     string
		language string
		gender   Gender
		age      uint8
		variant  uint8
	}

	isInit bool
}

func (ctx *Context) init() {
	if ctx.isInit {
		return
	}

	ctx.isInit = true
	ctx.rate = 175
	ctx.volume = 100
	ctx.pitch = 50
	ctx.tone = 50
}

func (ctx *Context) Rate() int {
	ctx.init()

	return ctx.rate
}

func (ctx *Context) Volume() int {
	ctx.init()

	return ctx.volume
}

func (ctx *Context) Pitch() int {
	ctx.init()

	return ctx.pitch
}

func (ctx *Context) Tone() int {
	ctx.init()

	return ctx.tone
}

func (ctx *Context) SetRate(wpm int) {
	if wpm < 80 || wpm > 450 {
		panic("espeak: Context.SetRate: wpm must be between 80 and 450")
	}

	ctx.init()

	ctx.rate = wpm
}

func (ctx *Context) SetVolume(percentage int) {
	if percentage < 0 {
		panic("espeak: Context.SetVolume: percentage must not be negative")
	}

	ctx.init()

	ctx.volume = percentage
}

func (ctx *Context) SetPitch(pitch int) {
	if pitch < 0 || pitch > 100 {
		panic("espeak: Context.SetPitch: pitch must be between 0 and 100")
	}

	ctx.init()

	ctx.pitch = pitch
}

func (ctx *Context) SetTone(tone int) {
	if tone < 0 || tone > 100 {
		panic("espeak: Context.SetTone: tone must be between 0 and 100")
	}

	ctx.init()

	ctx.tone = tone
}

type Voice struct {
	// a given name for this voice
	Name string

	// lists of pairs of priority + language (and dialect qualifier)
	Languages []Language

	// the filename for this voice within espeak-ng-data/voices
	Identifier string

	// gender of voice
	Gender Gender

	// age in years, or 0 if not specified
	Age uint8
}

type Language struct {
	// a low number indicates a more preferred voice, a higher number
	// indicates a less preferred voice.
	Priority uint8

	Name string
}

func ListVoices() []*Voice {
	lock.Lock()
	defer lock.Unlock()

	return listVoices()
}

type Gender uint8

const (
	Unknown Gender = 0
	Male    Gender = 1
	Female  Gender = 2
	Neutral Gender = 3
)

func (ctx *Context) SetVoice(name, language string) error {
	return ctx.SetVoiceProperties(name, language, Unknown, 0, 0)
}

func validVoice(name, language string, gender Gender, age, variant uint8) error {
	lock.Lock()
	defer lock.Unlock()

	return setVoice(name, language, gender, age, variant)
}

func (ctx *Context) SetVoiceProperties(name, language string, gender Gender, age, variant uint8) error {
	if err := validVoice(name, language, gender, age, variant); err != nil {
		return err
	}

	ctx.init()

	ctx.voice.name = name
	ctx.voice.language = language
	ctx.voice.gender = gender
	ctx.voice.age = age
	ctx.voice.variant = variant

	return nil
}

type SynthEventType uint8

const (
	// Start of word
	EventWord SynthEventType = 1

	// Start of sentence
	EventSentence SynthEventType = 2

	// Mark
	EventMark SynthEventType = 3

	// Audio event
	EventPlay SynthEventType = 4

	// End of sentence or clause
	EventEnd SynthEventType = 5

	// End of message
	EventMsgTerminated SynthEventType = 6

	// Phoneme, if enabled
	EventPhoneme SynthEventType = 7
)

type SynthEvent struct {
	Type SynthEventType

	// The number of characters from the start of the text
	TextPosition int

	// Word length, in characters (for EventWord)
	Length int

	// The time within the generated speech output data
	AudioPosition time.Duration

	Number  int    // used for EventWord and EventSentence
	Name    string // used for EventMark and EventPlay
	Phoneme string // used for EventPhoneme
}

// TODO:
/*
func (ctx *Context) Synthesize(speak *ssml.Speak) error {
	ctx.init()

	text, err := xml.Marshal(speak)
	if err != nil {
		return err
	}

	return ctx.synthesize(string(text))
}
*/

func (ctx *Context) SynthesizeText(text string) error {
	ctx.init()

	return ctx.synthesize(text)
}

func (ctx *Context) synthesize(text string) error {
	lock.Lock()
	defer lock.Unlock()

	if err := setRate(ctx.rate); err != nil {
		return err
	}

	if err := setVolume(ctx.volume); err != nil {
		return err
	}

	if err := setPitch(ctx.pitch); err != nil {
		return err
	}

	if err := setTone(ctx.tone); err != nil {
		return err
	}

	if err := setVoice(ctx.voice.name, ctx.voice.language, ctx.voice.gender, ctx.voice.age, ctx.voice.variant); err != nil {
		return err
	}

	return synthesize(text, ctx)
}
