// Package vad implements voice activated detection. This is achieved
// through reading the data with a buffer of 0.2 seconds. That means
// you cannot Read greater than 0.2 seconds worth of samples. If
// you need to be able to read more, then use a RealtimeStream on top of the
// VAD stream.
package vad

import (
	"math"

	"github.com/1lann/dissonance/audio"
)

// Filter represents a VAD filter.
type Filter struct {
	threshold float64
}

// streamFilter represents a stream filter.
type streamFilter struct {
	threshold     float64
	stream        audio.Stream
	buffer        []int32
	initialLength int
	taperOff      int
}

// NewFilter creates a new VAD filter with a threshold between 0 and 1.
func NewFilter(threshold float64) audio.Filter {
	if threshold < 0 || threshold > 1 {
		panic("vad: threshold must be between 0 and 1")
	}

	return &Filter{(threshold * threshold) * 2147483646}
}

// Filter implements the Filter method for filters.
func (f *Filter) Filter(stream audio.Stream) audio.Stream {
	return &streamFilter{
		threshold: f.threshold,
		stream:    stream,
		buffer:    make([]int32, stream.SampleRate()/5),
	}
}

func (f *streamFilter) Read(dst interface{}) (int, error) {
	dstLen := audio.SliceLength(dst)
	if dstLen > len(f.buffer) {
		return 0, audio.ErrBufferTooLarge
	}

	// Get average buffer
	if f.initialLength >= len(f.buffer) && f.getRMS() > f.threshold {
		// Return the buffer
		err := audio.ReadFromInt32(dst, f.buffer[:dstLen], dstLen)
		if err != nil {
			return 0, err
		}
		err = f.readToBuffer(dstLen)
		f.taperOff = len(f.buffer)
		return dstLen, err
	} else if f.taperOff > 0 {
		err := audio.ReadFromInt32(dst, f.buffer[:dstLen], dstLen)
		if err != nil {
			return 0, err
		}
		err = f.readToBuffer(dstLen)
		f.taperOff -= dstLen
		return dstLen, err
	} else if f.initialLength < len(f.buffer) {
		f.initialLength += dstLen
	}
	err := f.readToBuffer(dstLen)
	if err != nil {
		return 0, err
	}

	// Return silence
	err = audio.ReadFromInt32(dst, make([]int32, dstLen), dstLen)
	if err != nil {
		return 0, err
	}

	return dstLen, nil

}

func (f *streamFilter) SampleRate() int {
	return f.stream.SampleRate()
}

func (f *streamFilter) getRMS() float64 {
	var numPeaks float64
	var sum float64

	for i := 2; i < len(f.buffer); i++ {
		if f.buffer[i]-f.buffer[i-1] < 0 && f.buffer[i-1]-f.buffer[i-2] >= 0 ||
			f.buffer[i]-f.buffer[i-1] >= 0 && f.buffer[i-1]-f.buffer[i-2] < 0 {
			sum += math.Abs(float64(f.buffer[i-1])) / math.Sqrt2
			numPeaks++
		}
	}

	return sum / numPeaks
}

func (f *streamFilter) readToBuffer(num int) error {
	read := make([]int32, num)
	_, err := f.stream.Read(read)
	if err != nil {
		return err
	}

	f.buffer = append(f.buffer[num:], read...)

	return nil
}
