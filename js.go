// +build js

package espeak // import "gopkg.in/BenLubar/espeak.v2"

import (
	"time"

	"github.com/gopherjs/gopherjs/js"
)

var module = js.Global.Get("ESpeakNG").New()

func deref(ptr uintptr) uintptr {
	return uintptr(module.Call("getValue", ptr, "*").Int())
}

func getU8(ptr uintptr) uint8 {
	return uint8(module.Call("getValue", ptr, "i8").Int())
}

func getI16(ptr uintptr) int16 {
	return int16(module.Call("getValue", ptr, "i16").Int())
}

func getI32(ptr uintptr) int {
	return module.Call("getValue", ptr, "i32").Int()
}

func setPtr(ptr, val uintptr) {
	module.Call("setValue", ptr, val, "*")
}

func setU8(ptr uintptr, val uint8) {
	module.Call("setValue", ptr, val, "i8")
}

func malloc(size uintptr) uintptr {
	return uintptr(module.Call("_malloc", size).Int())
}

func free(ptr uintptr) {
	module.Call("_free", ptr)
}

func fromString(s string) uintptr {
	ptr := malloc(uintptr(len(s) + 1))

	for i := 0; i < len(s); i++ {
		setU8(ptr+uintptr(i), s[i])
	}
	setU8(ptr+uintptr(len(s)), 0)

	return ptr
}

func toString(ptr uintptr) string {
	var buf []byte

	b := getU8(ptr)
	for b != 0 {
		buf = append(buf, b)
		ptr++
		b = getU8(ptr)
	}

	return string(buf)
}

func toStringN(ptr uintptr, n int) string {
	var buf []byte

	b := getU8(ptr)
	for b != 0 {
		buf = append(buf, b)
		if len(buf) >= n {
			break
		}
		ptr++
		b = getU8(ptr)
	}

	return string(buf)
}

var errBuf = malloc(512)

func init() {
	lock.Lock()
	defer lock.Unlock()

	module.Call("_espeak_ng_InitializePath", 0)

	errCtx := malloc(4)
	defer free(errCtx)
	defer module.Call("_espeak_ng_ClearErrorContext", errCtx)

	status := module.Call("_espeak_ng_Initialize", errCtx)
	err := toErr(status)
	if err != nil {
		module.Call("_espeak_ng_PrintStatusCodeMessage", status, module.Call("_get_stderr"), deref(errCtx))
		panic(err)
	}

	status = module.Call("_espeak_ng_InitializeOutput", outputModeSynchronous, 0, 0)
	err = toErr(status)
	if err != nil {
		module.Call("_espeak_ng_PrintStatusCodeMessage", status, module.Get("stderr"), deref(errCtx))
		panic(err)
	}

	callback := module.Call("addFunction", synthCallback, "iiii")
	module.Call("_espeak_SetSynthCallback", callback)
}

func toErr(status *js.Object) error {
	if status.Int() == sOK {
		return nil
	}

	module.Call("_espeak_ng_GetStatusCodeMessage", status, errBuf, 512)
	return &Error{
		Code:    uint32(status.Int()),
		Message: toStringN(errBuf, 512),
	}
}

func getSampleRate() int {
	return module.Call("_espeak_ng_GetSampleRate").Int()
}

func listVoices() []*Voice {
	var voices []*Voice

	for cVoices := uintptr(module.Call("_espeak_ListVoices", 0).Int()); deref(cVoices) != 0; cVoices += 4 {
		voices = append(voices, toVoice(deref(cVoices)))
	}

	return voices
}

func toVoice(cVoice uintptr) *Voice {
	return &Voice{
		Name:       toString(cVoice + voiceNameOffset),
		Languages:  toLanguages(cVoice + voiceLanguagesOffset),
		Identifier: toString(cVoice + voiceIdentifierOffset),
		Gender:     Gender(getU8(cVoice + voiceGenderOffset)),
		Age:        getU8(cVoice + voiceAgeOffset),
	}
}

func toLanguages(data uintptr) []Language {
	var languages []Language

	for {
		priority := getU8(data)
		if priority == 0 {
			return languages
		}

		start, length, next := findNextLanguage(data)
		languages = append(languages, Language{
			Priority: priority,
			Name:     toStringN(start, length),
		})
		data = next
	}
}

func setRate(rate int) error {
	return toErr(module.Call("_espeak_ng_SetParameter", espeakRATE, rate, 0))
}

func setVolume(volume int) error {
	return toErr(module.Call("_espeak_ng_SetParameter", espeakVOLUME, volume, 0))
}

func setPitch(pitch int) error {
	return toErr(module.Call("_espeak_ng_SetParameter", espeakPITCH, pitch, 0))
}

func setTone(tone int) error {
	return toErr(module.Call("_espeak_ng_SetParameter", espeakRANGE, tone, 0))
}

func setVoice(name, language string, gender Gender, age, variant uint8) error {
	voice := malloc(voiceSize)
	defer free(voice)

	if name == "" {
		setPtr(voice+voiceNameOffset, 0)
	} else {
		cName := fromString(name)
		defer free(cName)
		setPtr(voice+voiceNameOffset, cName)
	}

	if name != "" && language == "" && gender == Unknown && age == 0 && variant == 0 {
		return toErr(module.Call("_espeak_ng_SetVoiceByName", deref(voice+voiceNameOffset)))
	}

	if language == "" {
		setPtr(voice+voiceLanguagesOffset, 0)
	} else {
		cLanguage := fromString(language)
		defer free(cLanguage)
		setPtr(voice+voiceLanguagesOffset, cLanguage)
	}

	setU8(voice+voiceGenderOffset, uint8(gender))
	setU8(voice+voiceAgeOffset, age)
	setU8(voice+voiceVariantOffset, variant)

	return toErr(module.Call("_espeak_ng_SetVoiceByProperties", voice))
}

var synthCtx *Context

func synthCallback(wav uintptr, numsamples int, events uintptr) int {
	for i, pwav := 0, wav; i < numsamples; i, pwav = i+1, pwav+2 {
		synthCtx.Samples = append(synthCtx.Samples, getI16(pwav))
	}

	for getI32(events+eventTypeOffset) != espeakEVENT_LIST_TERMINATED {
		if e := toEvent(events); e != nil {
			synthCtx.Events = append(synthCtx.Events, e)
		}

		events += eventSize
	}

	return 0 // continue synthesis
}

func toEvent(event uintptr) *SynthEvent {
	var synthEvent SynthEvent

	switch getI32(event + eventTypeOffset) {
	case espeakEVENT_WORD:
		synthEvent.Type = EventWord
		synthEvent.Number = getI32(event + eventNumberOffset)
	case espeakEVENT_SENTENCE:
		synthEvent.Type = EventSentence
		synthEvent.Number = getI32(event + eventNumberOffset)
	case espeakEVENT_MARK:
		synthEvent.Type = EventMark
		synthEvent.Name = toString(deref(event + eventNameOffset))
	case espeakEVENT_PLAY:
		synthEvent.Type = EventPlay
		synthEvent.Name = toString(deref(event + eventNameOffset))
	case espeakEVENT_END:
		synthEvent.Type = EventEnd
	case espeakEVENT_MSG_TERMINATED:
		synthEvent.Type = EventMsgTerminated
	case espeakEVENT_PHONEME:
		synthEvent.Type = EventPhoneme
		synthEvent.Phoneme = toStringN(event+eventStringOffset, 8)
	default:
		return nil
	}

	synthEvent.TextPosition = getI32(event + eventTextPositionOffset)
	synthEvent.Length = getI32(event + eventLengthOffset)
	synthEvent.AudioPosition = time.Duration(getI32(event+eventAudioPositionOffset)) * time.Millisecond

	return &synthEvent
}

func synthesize(text string, ctx *Context) error {
	synthCtx = ctx
	defer func() {
		synthCtx = nil
	}()

	cText := fromString(text)
	defer free(cText)

	return toErr(module.Call("_espeak_ng_Synthesize", cText, 0, 0, posCharacter, 0, espeakCHARS_UTF8|espeakSSML, 0, 0))
}

func findNextLanguage(data uintptr) (start uintptr, len int, next uintptr) {
	data++
	start = data
	for getU8(data) != 0 {
		data++
		len++
	}
	data++
	next = data

	return
}
