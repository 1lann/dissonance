// Package samplerate contains a sample rate converter using linear
// interpolation.
package samplerate

import "github.com/1lann/dissonance/audio"

// Filter represents the sample rate audio.Filter
type Filter struct {
	sampleRate int
}

type streamFilter struct {
	stream       audio.Stream
	sampleRate   int
	buffer       []int32
	lastPosition float64
	ratio        float64
}

// NewFilter returns a new sample rate filter to convert into the given
// sample rate.
func NewFilter(sampleRate int) audio.Filter {
	return &Filter{sampleRate: sampleRate}
}

func interpolate(buffer []int32, position float64) int32 {
	i := int(position)
	between := position - float64(int(position))
	return int32(float64(buffer[i]) + float64(buffer[i+1]-buffer[i])*between)
}

// Filter implements the Filter method for filters.
func (f *Filter) Filter(stream audio.Stream) audio.Stream {
	return &streamFilter{
		stream:       stream,
		sampleRate:   f.sampleRate,
		buffer:       []int32{},
		lastPosition: 0,
		ratio:        float64(stream.SampleRate()) / float64(f.sampleRate),
	}
}

func (f *streamFilter) SampleRate() int {
	return f.sampleRate
}

func (f *streamFilter) Read(dst interface{}) (int, error) {
	dstLen := audio.SliceLength(dst)
	required := int(float64(dstLen)*f.ratio+1) - len(f.buffer)
	buf := make([]int32, required)
	_, err := f.stream.Read(buf)
	if err != nil {
		return 0, err
	}

	f.buffer = append(f.buffer, buf...)

	var result []int32

	var i float64
	for ; i*f.ratio+f.lastPosition < float64(len(f.buffer)-1); i++ {
		result = append(result, interpolate(f.buffer, i*f.ratio+f.lastPosition))
	}

	i++
	f.lastPosition = i*f.ratio - float64(int(i*f.ratio))

	if f.lastPosition > 0 {
		f.buffer = f.buffer[len(f.buffer)-1:]
	} else {
		f.buffer = f.buffer[len(f.buffer):]
	}

	if err := audio.ReadFromInt32(dst, result, len(result)); err != nil {
		return 0, err
	}

	return len(result), nil
}
