package wav

import (
	"encoding/binary"
	"io"
	"os"
)

func startWav(w io.WriteSeeker, rate int) error {
	const (
		header1 = "RIFF\x24\xf0\xff\x7fWAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00"
		header2 = "\x02\x00\x10\x00data\x00\xf0\xff\x7f"
	)

	_, err := io.WriteString(w, header1)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, int32(rate))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, int32(rate*2))
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, header2)
	if err != nil {
		return err
	}

	return nil
}

func finishWav(w io.WriteSeeker) error {
	pos, err := w.Seek(0, os.SEEK_CUR)
	if err != nil {
		return err
	}

	_, err = w.Seek(4, os.SEEK_SET)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, int32(pos-8))
	if err != nil {
		return err
	}

	_, err = w.Seek(40, os.SEEK_SET)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, int32(pos-44))
	if err != nil {
		return err
	}

	return nil
}
