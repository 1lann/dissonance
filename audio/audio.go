package audio

import (
	"errors"
)

// ErrBufferTooLarge is returned when the audio buffer received is too large
// to process. This commonly occurs with streams for realtime applications.
var ErrBufferTooLarge = errors.New("audio: buffer is too large")

// DevicePair represents a pair of a playback and recording device.
type DevicePair struct {
	Playback PlaybackDevice
	Record   RecordingDevice
}

// Close closes both the playback and recording devices of a device pair.
func (a *DevicePair) Close() {
	a.Playback.Close()
	a.Record.Close()
}

// Stream represents an audio stream.
type Stream interface {
	SampleRate() int
	Read(interface{}) (int, error)
}

// PlaybackDevice represents a playback device that can play a stream.
type PlaybackDevice interface {
	PlayStream(Stream) error
	Close()
}

// RecordingDevice represents a recording device that can open a stream.
type RecordingDevice interface {
	OpenStream() (Stream, error)
	Close()
}

// Filter represents an audio stream filter.
type Filter interface {
	Filter(Stream) Stream
}
