package wav

import (
	"io"

	"github.com/BenLubar/espeak"
)

type Concat struct {
	buf buffer
}

func (c *Concat) Synth(voice *espeak.Voice, text string, pitch, pitchRange, rate int) error {
	lock.Lock()

	if voice != nil {
		if err := espeak.SetVoiceByName(voice.Name); err != nil {
			lock.Unlock()
			return err
		}
	}

	if err := espeak.SetPitch(pitch); err != nil {
		lock.Unlock()
		return err
	}
	if err := espeak.SetRange(pitchRange); err != nil {
		lock.Unlock()
		return err
	}
	if err := espeak.SetRate(rate); err != nil {
		lock.Unlock()
		return err
	}

	var data userData
	data.done = make(chan error, 1)

	if err := espeak.Synth(text, 0, espeak.PosCharacter, 0, 0, nil, &data); err != nil {
		lock.Unlock()
		return err
	}
	lock.Unlock()

	if err := <-data.done; err != nil {
		return err
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
