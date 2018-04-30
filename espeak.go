// Package espeak provides bindings for eSpeak <http://espeak.sf.net>.
//
// Please note that functions in this package should only be called from one
// goroutine at a time as libespeak uses static variables for most of its
// returned values.
package espeak // import "gopkg.in/BenLubar/espeak.v1"

/*
#cgo LDFLAGS: -lespeak
#include <stdlib.h> // for C.free
#include "helper.h" // also pulls in <espeak/speak_lib.h>

int synthCallback(short *, int, espeak_EVENT *);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type AudioOutput C.espeak_AUDIO_OUTPUT

const (
	// PLAYBACK mode: plays the audio data, supplies events to the calling
	// program.
	AudioOutputPlayback = AudioOutput(C.AUDIO_OUTPUT_PLAYBACK)

	// RETRIEVAL mode: supplies audio data and events to the calling
	// program.
	AudioOutputRetrieval = AudioOutput(C.AUDIO_OUTPUT_RETRIEVAL)

	// SYNCHRONOUS mode: as RETRIEVAL but doesn't return until synthesis is
	// completed.
	AudioOutputSynchronous = AudioOutput(C.AUDIO_OUTPUT_SYNCHRONOUS)

	// Synchronous playback.
	AudioOutputSynchPlayback = AudioOutput(C.AUDIO_OUTPUT_SYNCH_PLAYBACK)
)

const (
	InitializePhonemeEvents = 1 << 0
	InitializePhonemeIPA    = 1 << 1
	InitializeDontExit      = 1 << 15
)

var audioOutputRetrieval bool

// Initialize must be called before any synthesis functions are called.
//
// output: The audio data can either be played by eSpeak or passed back by the
//         argument to the SetSynthCallback function.
// buflength: The length in milliseconds of sound buffers passed to the
//            argument to the SetSynthCallback functon. 0 gives the default
//            value of 200ms. This paramater is only used for
//            AudioOutputRetrieval and AudioOutputSynchronous modes.
// path: The directory which contains the espeak-data directory, or an empty
//       string for the default location.
// options: These may be OR'd together:
//          InitializePhonemeEvents:  Allow espeak.PhonemeEvent events.
//          InitializePhonemeIPA:     espeak.PhonemeEvent events give IPA
//                                    phoneme names, not eSpeak phoneme names.
//          InitializeDontExit:       Don't exit if espeak-data is not found.
//                                    (used for --help)
//
// Returns: sample rate in Hz, or ErrInternalError.
func Initialize(output AudioOutput, buflength int, path string, options int) (sampleRate int, err error) {
	var cpath *C.char
	if path != "" {
		cpath = C.CString(path)
		defer C.free(unsafe.Pointer(cpath))
	}

	switch output {
	case AudioOutputPlayback, AudioOutputSynchPlayback:
		audioOutputRetrieval = false
	case AudioOutputRetrieval, AudioOutputSynchronous:
		audioOutputRetrieval = true
	default:
		panic(fmt.Sprintf("espeak: unexpected value for output: %v", output))
	}

	sampleRate = int(C.espeak_Initialize(C.espeak_AUDIO_OUTPUT(output), C.int(buflength), cpath, C.int(options)))
	if sampleRate == -1 {
		sampleRate, err = 0, ErrInternalError
	}

	return
}

var (
	ErrInternalError = errors.New("espeak: internal error")
	ErrBufferFull    = errors.New("espeak: buffer full")
	ErrNotFound      = errors.New("espeak: not found")
)

func espeakError(err C.espeak_ERROR) error {
	switch err {
	case C.EE_OK:
		return nil

	case C.EE_INTERNAL_ERROR:
		return ErrInternalError

	case C.EE_BUFFER_FULL:
		return ErrBufferFull

	case C.EE_NOT_FOUND:
		return ErrNotFound

	default:
		panic(fmt.Sprintf("espeak: unknown error code %d", int(err)))
	}
}

type PositionType C.espeak_POSITION_TYPE

const (
	PosCharacter = PositionType(C.POS_CHARACTER)
	PosWord      = PositionType(C.POS_WORD)
	PosSentence  = PositionType(C.POS_SENTENCE)
)

const (
	SynthSSML     = uint(C.espeakSSML)
	SynthPhonemes = uint(C.espeakPHONEMES)
	SynthEndPause = uint(C.espeakENDPAUSE)
)

// Synthesize speech for the specified text. The speech sound data is passed
// to the calling program in buffers by means of the callback function
// specified by espeak.SetSynthCallback. The command is asynchronous: it is
// internally buffered and returns as soon as possible. If espeak.Initialize
// was previously called with AudioOutputPlayback as argument, the sound data
// are played by eSpeak.
//
// text: The text to be spoken.
//
// position:  The position in the text where speaking starts. Zero indicates
//            speak from the start of the text.
//
// positionType:  Determines whether "position" is a number of characters,
//                words, or sentences.
//
// endPosition:  If set, this gives a character position at which speaking will
//               stop.  A value of zero indicates no end position.
//
// flags:  These may be OR'd together:
//    SynthSSML      Elements within < > are treated as SSML elements, or if
//                   not recognised are ignored.
//
//    SynthPhonemes  Text within [[ ]] is treated as phonemes codes (in
//                   espeak's Hirshenbaum encoding).
//
//    SynthEndPause  If set then a sentence pause is added at the end of the
//                   text.  If not set then this pause is suppressed.
//
// uniqueIdentifier: an integer variable to which eSpeak writes a message
//                   identifier number, or nil. eSpeak includes this number in
//                   Event messages which are the result of this call
//
// userData: an interface{} or nil, which will be passed to the callback
//           function in Event messages.
//
// Errors:
//         ErrBufferFull: the command cannot be buffered; you may try after a
//                        while to call the function again.
//         ErrInternalError.
func Synth(text string, position uint, positionType PositionType, endPosition, flags uint, uniqueIdentifier *uint, userData interface{}) error {
	var cuserData unsafe.Pointer
	if userData != nil {
		ptr := &userData
		gcHelper[ptr] = struct{}{}
		cuserData = unsafe.Pointer(ptr)
	}
	cflags := C.uint(flags)&^7 | C.espeakCHARS_UTF8
	csize := C.size_t(len(text) + 1)
	ctext := unsafe.Pointer(C.CString(text))
	defer C.free(ctext)

	var cuniqueIdentifier *C.uint
	if uniqueIdentifier != nil {
		cuniqueIdentifier = new(C.uint)
	}

	err := espeakError(C.espeak_Synth(ctext, csize, C.uint(position), C.espeak_POSITION_TYPE(positionType), C.uint(endPosition), cflags, cuniqueIdentifier, cuserData))
	if err != nil && userData != nil {
		// XXX: does an error being returned guarantee that the
		// user_data parameter is discarded?
		delete(gcHelper, &userData)
	}

	if uniqueIdentifier != nil {
		*uniqueIdentifier = uint(*cuniqueIdentifier)
	}

	return err
}

var synthCallback0 func([]int16, []Event) bool

func SetSynthCallback(f func(wav []int16, events []Event) (stop bool)) {
	synthCallback0 = f
	C.espeak_SetSynthCallback((*C.t_espeak_callback)(C.synthCallback))
}

//export synthCallback
func synthCallback(cwav *C.short, numsamples C.int, cevents *C.espeak_EVENT) C.int {
	var wav []int16
	if cwav != nil {
		wav = make([]int16, numsamples)
		for i := range wav {
			// complicated way of saying wav[i] = cwav[i]
			wav[i] = (int16)(*(*C.short)(unsafe.Pointer(uintptr(unsafe.Pointer(cwav)) + unsafe.Sizeof(*cwav)*uintptr(i))))
		}
	}

	var events []Event
	for {
		events = append(events, parseEvent(cevents))

		if C.eventType(cevents) == C.espeakEVENT_LIST_TERMINATED {
			break
		}

		cevents = (*C.espeak_EVENT)(unsafe.Pointer(uintptr(unsafe.Pointer(cevents)) + unsafe.Sizeof(*cevents)))

		// playback modes don't get multiple events per call.
		if !audioOutputRetrieval {
			break
		}
	}

	if synthCallback0(wav, events) {
		return 1
	}
	return 0
}

// The C variables can't be seen by the garbage collector, so we store a copy
// of the pointer here.
var gcHelper = make(map[*interface{}]struct{})

func parseEvent(cevent *C.espeak_EVENT) Event {
	var data interface{}
	if cevent.user_data != nil {
		ptr := (*interface{})(cevent.user_data)
		data = *ptr
		if C.eventType(cevent) == C.espeakEVENT_MSG_TERMINATED {
			// We don't need it anymore.
			delete(gcHelper, ptr)
		}
	}
	base := BaseEvent{
		UniqueIdentifier: uint(cevent.unique_identifier),
		TextPosition:     int(cevent.text_position),
		Length:           int(cevent.length),
		AudioPosition:    int(cevent.audio_position),
		Sample:           int(cevent.sample),
		UserData:         data,
	}
	switch C.eventType(cevent) {
	case C.espeakEVENT_WORD:
		return WordEvent{
			BaseEvent: base,
			Number:    int(C.eventNumber(cevent)),
		}
	case C.espeakEVENT_SENTENCE:
		return SentenceEvent{
			BaseEvent: base,
			Number:    int(C.eventNumber(cevent)),
		}
	case C.espeakEVENT_MARK:
		return MarkEvent{
			BaseEvent: base,
			Name:      C.GoString(C.eventName(cevent)),
		}
	case C.espeakEVENT_PLAY:
		return PlayEvent{
			BaseEvent: base,
			Name:      C.GoString(C.eventName(cevent)),
		}
	case C.espeakEVENT_END:
		return EndEvent{
			BaseEvent: base,
		}
	case C.espeakEVENT_MSG_TERMINATED:
		return MsgTerminatedEvent{
			BaseEvent: base,
		}
	case C.espeakEVENT_PHONEME:
		return PhonemeEvent{
			BaseEvent: base,
			String:    C.GoString(C.eventString(cevent)),
		}
	case C.espeakEVENT_SAMPLERATE:
		return SampleRateEvent{
			BaseEvent: base,
		}
	case C.espeakEVENT_LIST_TERMINATED:
		return ListTerminatedEvent{
			BaseEvent: base,
		}
	default:
		panic(fmt.Sprintf("espeak: unexpected event type %d", C.eventType(cevent)))
	}
}

type (
	Event interface {
		isEvent()
	}

	BaseEvent struct {
		UniqueIdentifier uint
		TextPosition     int
		Length           int
		AudioPosition    int
		Sample           int
		UserData         interface{}
	}

	// Start of word
	WordEvent struct {
		BaseEvent
		Number int
	}
	// Start of sentence
	SentenceEvent struct {
		BaseEvent
		Number int
	}
	// <mark> element
	MarkEvent struct {
		BaseEvent
		Name string
	}
	// <audio> element
	PlayEvent struct {
		BaseEvent
		Name string
	}
	// End of sentence or clause
	EndEvent struct {
		BaseEvent
	}
	// End of message
	MsgTerminatedEvent struct {
		BaseEvent
	}
	// Phoneme, if enabled in espeak.Initialize()
	PhonemeEvent struct {
		BaseEvent
		String string
	}
	// Set sample rate
	SampleRateEvent struct {
		BaseEvent
	}
	// Retrieval mode: terminates the event list.
	ListTerminatedEvent struct {
		BaseEvent
	}
)

func (BaseEvent) isEvent() {}

var uriCallback0 func(string, string) bool

func SetURICallback(f func(uri, base string) (replaceWithText bool)) {
	uriCallback0 = f
	C.uriSetCallback()
}

//export uriCallback
func uriCallback(curi, cbase *C.char) C.int {
	if uriCallback0(C.GoString(curi), C.GoString(cbase)) {
		return 1
	}
	return 0
}

type Gender uint8

const (
	NoGender Gender = iota
	Male
	Female
)

type Voice struct {
	Name      string // a given name for this voice
	Languages []struct {
		Priority uint8  // lower is more preferred
		Language string // language code with optional dialect
	}
	Identifier string // the filename for this voice
	Gender     Gender
	Age        uint8 // 0 = not specified, or the age in years
}

func espeakVoice(cvoice *C.espeak_VOICE) *Voice {
	if cvoice == nil || *cvoice == (C.espeak_VOICE{}) {
		return nil
	}

	voice := &Voice{
		Name:       C.GoString(cvoice.name),
		Identifier: C.GoString(cvoice.identifier),
		Gender:     Gender(cvoice.gender),
		Age:        uint8(cvoice.age),
	}

	lang := cvoice.languages
	if lang != nil {
		for {
			priority := uint8(*lang)
			if priority == 0 {
				break
			}
			name := C.GoString((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(lang)) + 1)))

			voice.Languages = append(voice.Languages, struct {
				Priority uint8
				Language string
			}{
				Priority: priority,
				Language: name,
			})

			lang = (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(lang)) + 1 + uintptr(len(name)) + 1))
		}
	}

	return voice
}

// Returns the *espeak.Voice data for the currently selected voice. This is not
// affected by temporary voice changes caused by SSML elements such as <voice>
// and <s>.
func CurrentVoice() *Voice {
	return espeakVoice(C.espeak_GetCurrentVoice())
}

// Reads the voice files from espeak-data/voices and creates a slice of
// *espeak.Voice. All voices are listed in an undefined order.
func ListVoices() []*Voice {
	var voices []*Voice

	for cvoices := C.espeak_ListVoices(nil); *cvoices != nil; cvoices = (**C.espeak_VOICE)(unsafe.Pointer(uintptr(unsafe.Pointer(cvoices)) + unsafe.Sizeof(*cvoices))) {
		voices = append(voices, espeakVoice(*cvoices))
	}

	return voices
}

// Reads the voice files from espeak-data/voices and creates a slice of
// *espeak.Voice. Matching voices are listed in preference order.
func ListVoicesByProperties(name, language string, gender Gender, age uint8) []*Voice {
	var cname, clanguage *C.char
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	if language != "" {
		clanguage = C.CString(language)
		defer C.free(unsafe.Pointer(clanguage))
	}
	spec := C.espeak_VOICE{
		name:      cname,
		languages: clanguage,
		gender:    C.uchar(gender),
		age:       C.uchar(age),
	}

	var voices []*Voice

	for cvoices := C.espeak_ListVoices(&spec); *cvoices != nil; cvoices = (**C.espeak_VOICE)(unsafe.Pointer(uintptr(unsafe.Pointer(cvoices)) + unsafe.Sizeof(*cvoices))) {
		voices = append(voices, espeakVoice(*cvoices))
	}

	return voices
}

// Searches for a voice with a matching "name" field. Language is not
// considered.
func SetVoiceByName(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return espeakError(C.espeak_SetVoiceByName(cname))
}

// Sets the voice to one that matches the arguments.
//
// name: An empty string or a voice name.
// language: An empty string or a single language string (with optional
//           dialect), eg. "en-uk" or "en".
// gender: NoGender (any) or Male/Female.
// age: 0 or an age in years.
// variant: The index in the slice that ListVoicesByProperties would return for
//          the same arguments. 0 for the best match.
func SetVoiceByProperties(name, language string, gender Gender, age uint8, variant int) error {
	var cname, clanguage *C.char
	if name != "" {
		cname = C.CString(name)
		defer C.free(unsafe.Pointer(cname))
	}
	if language != "" {
		clanguage = C.CString(language)
		defer C.free(unsafe.Pointer(clanguage))
	}
	spec := C.espeak_VOICE{
		name:      cname,
		languages: clanguage,
		gender:    C.uchar(gender),
		age:       C.uchar(age),
		variant:   C.uchar(variant),
	}
	return espeakError(C.espeak_SetVoiceByProperties(&spec))
}

// Rate of speech in words per minute, [80, 450].
func Rate() int {
	return int(C.espeak_GetParameter(C.espeakRATE, 1))
}
func SetRate(wpm int) error {
	return espeakError(C.espeak_SetParameter(C.espeakRATE, C.int(wpm), 0))
}

// Volume, [0, ∞). 0 = silent, 100 = normal full volume, >100 may be distorted.
func Volume() int {
	return int(C.espeak_GetParameter(C.espeakVOLUME, 1))
}
func SetVolume(percent int) error {
	return espeakError(C.espeak_SetParameter(C.espeakVOLUME, C.int(percent), 0))
}

// Pitch base, [0, 100]. 50 = normal.
func Pitch() int {
	return int(C.espeak_GetParameter(C.espeakPITCH, 1))
}
func SetPitch(percent int) error {
	return espeakError(C.espeak_SetParameter(C.espeakPITCH, C.int(percent), 0))
}

// Range of pitch, [0, 100]. 0 = monotone, 50 = normal.
func Range() int {
	return int(C.espeak_GetParameter(C.espeakRANGE, 1))
}
func SetRange(percent int) error {
	return espeakError(C.espeak_SetParameter(C.espeakRANGE, C.int(percent), 0))
}

func Synchronize() error {
	return espeakError(C.espeak_Synchronize())
}

func Cancel() error {
	return espeakError(C.espeak_Cancel())
}

type PhonemeMode int

const (
	PhonemeOnly       PhonemeMode = 0 // just phonemes
	PhonemeTies       PhonemeMode = 1 // include ties (U+0361) for multi-letter names
	PhonemeZWJ        PhonemeMode = 2 // include zero-width joiners for phoneme names of more than one letter
	PhonemeUnderscore PhonemeMode = 3 // separate phonemes with underscore characters
)

// TextToPhonemes translates text into phonemes.  Call SetVoiceByName() first,
// to select a language.
//
// It returns a string which contains the phonemes for the text up to
// end of a sentence, or comma, semicolon, colon, or similar punctuation,
// as well as a string containing the remaining characters.
func TextToPhonemes(text string, phonemeMode PhonemeMode, ipa bool) (phonemes, remaining string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	cRemaining := unsafe.Pointer(cText)

	if ipa {
		phonemeMode |= 1 << 4
	}

	cPhonemes := C.espeak_TextToPhonemes(&cRemaining, C.espeakCHARS_UTF8, C.int(phonemeMode))
	phonemes = C.GoString(cPhonemes)
	if cRemaining != nil {
		remaining = C.GoString((*C.char)(cRemaining))
	}
	return
}
