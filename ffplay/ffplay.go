package ffplay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/1lann/dissonance/audio"
)

// SampleRate is the sample rate used by FFMPEG.
const SampleRate = 48000

type FFPlaySink struct {
	cmd   *exec.Cmd
	debug bool
}

func (f *FFPlaySink) Close() {
	if f.cmd != nil && f.cmd.Process != nil {
		f.cmd.Process.Kill()
	}
}

func (f *FFPlaySink) PlayStream(stream audio.Stream) error {
	if stream.SampleRate() != SampleRate {
		return errors.New("ffplay: sample rate must be 48000")
	}

	stderr, err := f.cmd.StderrPipe()
	if err != nil {
		return err
	}

	wr, err := f.cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = f.cmd.Start()
	if err != nil {
		return err
	}

	if f.debug {
		go io.Copy(os.Stderr, stderr)
	} else {
		go io.Copy(ioutil.Discard, stderr)
	}

	go func() {
		f.cmd.Wait()
		stderr.Close()
		wr.Close()
	}()

	buffer := make([]int32, 2400)
	writeBuffer := new(bytes.Buffer)

	for {
		n, err := stream.Read(buffer)
		if err != nil {
			return err
		}

		for _, sample := range buffer[:n] {
			binary.Write(writeBuffer, binary.LittleEndian, sample)
		}

		_, err = writeBuffer.WriteTo(wr)
		if err != nil {
			return err
		}
	}
}

// NewFFPlaySink creates a ffplay audio sink.
func NewFFPlaySink(debug ...bool) audio.PlaybackDevice {
	shouldDebug := false
	if len(debug) > 0 && debug[0] == true {
		shouldDebug = true
	}

	return &FFPlaySink{
		cmd:   exec.Command("ffplay", "-f", "pcm_s32le", "-ar", "48000", "-ac", "1", "-"),
		debug: shouldDebug,
	}
}
