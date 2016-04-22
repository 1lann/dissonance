package audio

import (
	"time"
)

type realtimeStream struct {
	stream       Stream
	lastError    error
	buffer       []int32
	readPosition int
}

// NewRealtimeStream converts a stream into a buffered stream for use in
// realtime applications, such as audio over a network.
func NewRealtimeStream(stream Stream) Stream {
	newStream := &realtimeStream{
		stream:       stream,
		lastError:    nil,
		buffer:       make([]int32, 4096),
		readPosition: 4096,
	}

	go newStream.run()
	return newStream
}

func (r *realtimeStream) run() {
	buffer := make([]int32, 1024)
	for {
		n, err := r.stream.Read(buffer)
		r.buffer = append(r.buffer[:len(r.buffer)-n], buffer[:n]...)

		r.readPosition -= n
		if r.readPosition < 0 {
			r.readPosition = len(r.buffer) / 2
		}

		if err != nil {
			r.lastError = err
			return
		}
	}
}

func (r *realtimeStream) Read(dst interface{}) (int, error) {
	dstLen := SliceLength(dst)
	if dstLen > 2048 {
		return 0, ErrBufferTooLarge
	}

	for {
		available := (4096 - r.readPosition)
		if available >= dstLen || r.lastError != nil {
			if dstLen > available {
				err := ReadFromInt32(dst, r.buffer[r.readPosition:], available)
				r.readPosition += available
				if r.lastError != nil {
					return available, r.lastError
				}

				return available, err
			} else {
				err := ReadFromInt32(dst, r.buffer[r.readPosition:], dstLen)
				r.readPosition += dstLen
				if r.lastError != nil {
					return available, r.lastError
				}

				return available, err
			}
		} else {
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (r *realtimeStream) SampleRate() int {
	return r.stream.SampleRate()
}
