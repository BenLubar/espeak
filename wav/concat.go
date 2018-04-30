package wav // import "gopkg.in/BenLubar/espeak.v1/wav"

import (
	"io"
	"time"

	"github.com/BenLubar/espeak"
)

type Concat struct {
	buf buffer
}

func (c *Concat) Synth(voice *espeak.Voice, text string, pitch, pitchRange, rate int) error {
	lock.Lock()
	defer lock.Unlock()

	if voice != nil {
		if err := espeak.SetVoiceByName(voice.Name); err != nil {
			return err
		}
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

	var data userData
	data.done = make(chan error, 1)

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

	_, err := data.buf.WriteTo(&c.buf)

	return err
}

func (c *Concat) WriteTo(w io.Writer) (n int64, err error) {
	var buf buffer

	if err := startWav(&buf, sampleRate); err != nil {
		return 0, err
	}

	if _, err := c.buf.WriteTo(&buf); err != nil {
		return 0, err
	}

	if err := finishWav(&buf); err != nil {
		return 0, err
	}

	return buf.WriteTo(w)
}
