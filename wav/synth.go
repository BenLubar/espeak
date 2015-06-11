package wav

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/BenLubar/espeak"
)

var ErrTimeout = errors.New("wav: timeout")

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

		if err := binary.Write(&data.buf, binary.LittleEndian, wav); err != nil {
			data.done <- err
			return true
		}

		if _, ok := events[0].(espeak.MsgTerminatedEvent); ok {
			data.done <- nil
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

func Synth(w io.Writer, voice *espeak.Voice, text string, pitch, pitchRange, rate int) error {
	lock.Lock()
	defer lock.Unlock()

	if voice != nil {
		if err := espeak.SetVoiceByName(voice.Name); err != nil {
			return err
		}
	}

	var data userData
	data.done = make(chan error, 1)

	if err := startWav(&data.buf, sampleRate); err != nil {
		return err
	}

	if err := espeak.SetPitch(pitch); err != nil {
		return err
	}
	if err := espeak.SetRange(pitchRange); err != nil {
		return err
	}
	if err := espeak.SetRate(rate); err != nil {
		return err
	}

	if err := espeak.Synth(text, 0, espeak.PosCharacter, 0, 0, nil, &data); err != nil {
		return err
	}

	select {
	case err := <-data.done:
		if err != nil {
			return err
		}

	case <-time.After(time.Second):
		if err := espeak.Cancel(); err != nil {
			return err
		}

		return ErrTimeout
	}

	if err := finishWav(&data.buf); err != nil {
		return err
	}

	_, err := data.buf.WriteTo(w)
	return err
}
