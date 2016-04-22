package paudio

import (
	"github.com/1lann/dissonance/audio"
	"github.com/gordonklaus/portaudio"
)

const bufferSize = 512

var hasInitialized = false

// NewPlaybackDevice returns the default portaudio playback device.
func NewPlaybackDevice() (audio.PlaybackDevice, error) {
	if !hasInitialized {
		if err := portaudio.Initialize(); err != nil {
			return nil, err
		}
		hasInitialized = true
	}

	return &playbackDevice{
		internalStream: nil,
		buffer:         make([]int32, bufferSize),
	}, nil
}

// NewRecordingDevice returns the default portaudio recording device.
func NewRecordingDevice() (audio.RecordingDevice, error) {
	if !hasInitialized {
		if err := portaudio.Initialize(); err != nil {
			return nil, err
		}
		hasInitialized = true
	}

	return new(recordingDevice), nil
}

type playbackDevice struct {
	internalStream *portaudio.Stream
	buffer         []int32
}

func (d *playbackDevice) PlayStream(stream audio.Stream) error {
	var err error
	d.internalStream, err = portaudio.OpenDefaultStream(
		0, 1, float64(stream.SampleRate()), len(d.buffer), &d.buffer)
	if err != nil {
		return err
	}

	err = d.internalStream.Start()

	for {
		_, err = stream.Read(d.buffer)
		if err != nil {
			return err
		}

		d.internalStream.Write()
	}
}

func (d *playbackDevice) Close() {
	d.internalStream.Stop()
	d.internalStream.Close()

}

type recordingDevice struct {
	internalStream *portaudio.Stream
}

type recordingStream struct {
	internalStream *portaudio.Stream
	buffer         []int32
	sampleRate     int
	buffered       []int32
	started        bool
}

func (d *recordingDevice) OpenStream() (audio.Stream, error) {
	stream := &recordingStream{
		internalStream: nil,
		buffer:         make([]int32, bufferSize),
		sampleRate:     44100,
		buffered:       []int32{},
		started:        false,
	}

	var err error
	stream.internalStream, err = portaudio.OpenDefaultStream(1, 0,
		float64(stream.sampleRate), len(stream.buffer), &stream.buffer)
	if err != nil {
		return nil, err
	}
	d.internalStream = stream.internalStream

	return stream, nil
}

func (d *recordingDevice) Close() {
	d.internalStream.Stop()
	d.internalStream.Close()
}

func (s *recordingStream) SampleRate() int {
	return s.sampleRate
}

func (s *recordingStream) Read(dst interface{}) (int, error) {
	if !s.started {
		s.started = true
		err := s.internalStream.Start()
		if err != nil {
			return 0, err
		}
	}

	dstLen := audio.SliceLength(dst)
	for len(s.buffered) < dstLen {
		err := s.internalStream.Read()
		if err != nil {
			return 0, err
		}
		s.buffered = append(s.buffered, s.buffer...)
	}

	err := audio.ReadFromInt32(dst, s.buffered, dstLen)
	s.buffered = s.buffered[dstLen:]
	return dstLen, err
}
