package ffmpeg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/1lann/dissonance/audio"
)

// SampleRate is the sample rate used by FFMPEG.
const SampleRate = 48000

func newFFMPEGStream(cmd *exec.Cmd, debug bool) (audio.Stream, error) {
	outRd, outWr := io.Pipe()
	errRd, errWr := io.Pipe()
	cmd.Stdout = outWr
	cmd.Stderr = errWr

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	if debug {
		go io.Copy(os.Stderr, errRd)
	} else {
		go io.Copy(ioutil.Discard, errRd)
	}

	go func() {
		cmd.Wait()
		errWr.Close()
		outWr.Close()
	}()

	b := make([]byte, 1)
	_, err = outRd.Read(b)
	if err != nil {
		return nil, errors.New("ffmpeg: failed to start, enable debug to view details")
	}

	stream := audio.NewOfflineStream(SampleRate, SampleRate)

	go stream.ReadBytes(io.MultiReader(bytes.NewReader(b), outRd),
		binary.LittleEndian, audio.Int32)

	return stream, nil
}

// NewFFMPEGStream returns an audio stream from any input that FFMPEG accepts.
func NewFFMPEGStream(input io.Reader, debug ...bool) (audio.Stream, error) {
	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-acodec", "pcm_s32le",
		"-f", "s32le", "-ac", "1", "-ar", strconv.Itoa(SampleRate), "pipe:1")
	cmd.Stdin = input
	return newFFMPEGStream(cmd, len(debug) > 0 && debug[0])
}

// NewFFMPEGStreamFromFile returns an audio stream from the given filename.
func NewFFMPEGStreamFromFile(name string, debug ...bool) (audio.Stream, error) {
	cmd := exec.Command("ffmpeg", "-i", name, "-acodec", "pcm_s32le",
		"-f", "s32le", "-ac", "1", "-ar", strconv.Itoa(SampleRate), "pipe:1")
	return newFFMPEGStream(cmd, len(debug) > 0 && debug[0])
}

var devicePattern = regexp.MustCompile(`\[dshow @ [0-9a-f]+\]  "(.+)"`)

func GetDshowDevices() ([]string, error) {
	cmd := exec.Command("ffmpeg", "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	output, err := cmd.CombinedOutput()
	if cmd.ProcessState.ExitCode() != 1 && err != nil {
		return nil, fmt.Errorf("ffmpeg: error invoking ffmpeg: %w\n%s", err, string(output))
	}

	results := devicePattern.FindAllStringSubmatch(string(output), -1)

	out := make([]string, len(results))
	for i, res := range results {
		out[i] = res[1]
	}

	return out, nil
}

func NewFFMPEGStreamFromDshow(device string, debug ...bool) (audio.Stream, error) {
	cmd := exec.Command("ffmpeg", "-f", "dshow", "-audio_buffer_size", "50", "-i", `audio=`+device, "-acodec", "pcm_s32le",
		"-f", "s32le", "-ac", "1", "-ar", strconv.Itoa(SampleRate), "pipe:1")
	return newFFMPEGStream(cmd, len(debug) > 0 && debug[0])
}
