package wav

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/BenLubar/espeak"
)

var lock sync.Mutex

var sampleRate int

func init() {
	var err error
	sampleRate, err = espeak.Initialize(espeak.AudioOutputRetrieval, 0, "", 0)
	if err != nil {
		panic(err)
	}

	espeak.SetSynthCallback(func(wav []int16, events []espeak.Event) (stop bool) {
		data := events[len(events)-1].(espeak.ListTerminatedEvent).UserData.(*userData)
		if wav == nil {
			data.done <- nil
			return false
		}

		if err := binary.Write(&data.buf, binary.LittleEndian, wav); err != nil {
			data.done <- err
			return true
		}

		return false
	})
}

func AllVoices() []*espeak.Voice {
	lock.Lock()
	defer lock.Unlock()

	return espeak.ListVoices()
}

func FindVoices(name, language string, gender espeak.Gender, age uint8) []*espeak.Voice {
	lock.Lock()
	defer lock.Unlock()

	return espeak.ListVoicesByProperties(name, language, gender, age)
}

type userData struct {
	done chan error
	buf  buffer
}

func Synth(w io.Writer, voice *espeak.Voice, text string) error {
	lock.Lock()

	if voice != nil {
		if err := espeak.SetVoiceByName(voice.Name); err != nil {
			lock.Unlock()
			return err
		}
	}

	var data userData
	data.done = make(chan error, 1)

	if err := startWav(&data.buf, sampleRate); err != nil {
		lock.Unlock()
		return err
	}

	if err := espeak.Synth(text, 0, espeak.PosCharacter, 0, 0, nil, &data); err != nil {
		lock.Unlock()
		return err
	}
	lock.Unlock()

	if err := <-data.done; err != nil {
		return err
	}

	if err := finishWav(&data.buf); err != nil {
		return err
	}

	_, err := data.buf.WriteTo(w)
	return err
}
