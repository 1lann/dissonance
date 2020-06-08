package audio

import (
	"sync"
	"time"
)

type realtimeStream struct {
	stream       Stream
	lastError    error
	buffer       []int32
	readPosition int
	usageLock    *sync.Mutex
}

// NewRealtimeStream converts a stream into a buffered stream for use in
// realtime applications, such as audio over a network.
func NewRealtimeStream(stream Stream, bufferSize int) Stream {
	newStream := &realtimeStream{
		stream:       stream,
		lastError:    nil,
		buffer:       make([]int32, bufferSize),
		readPosition: bufferSize,
		usageLock:    new(sync.Mutex),
	}

	go newStream.run()
	return newStream
}

func (r *realtimeStream) run() {
	buffer := make([]int32, 1024)
	for {
		n, err := r.stream.Read(buffer)
		r.usageLock.Lock()
		r.buffer = append(r.buffer[n:], buffer[:n]...)

		r.readPosition -= n
		if r.readPosition < 0 {
			r.readPosition = len(r.buffer) / 2
		}

		if err != nil {
			r.lastError = err
			r.usageLock.Unlock()
			return
		}
		r.usageLock.Unlock()
	}
}

func (r *realtimeStream) Read(dst interface{}) (int, error) {
	dstLen := SliceLength(dst)
	r.usageLock.Lock()
	if dstLen > len(r.buffer)/2 {
		dstLen = len(r.buffer) / 2
	}
	r.usageLock.Unlock()

	for {
		r.usageLock.Lock()
		available := (len(r.buffer) - r.readPosition)
		if available >= dstLen || r.lastError != nil {
			if dstLen > available {
				err := ReadFromInt32(dst, r.buffer[r.readPosition:], available)
				r.readPosition += available
				if r.lastError != nil {
					r.usageLock.Unlock()
					return available, r.lastError
				}

				r.usageLock.Unlock()
				return available, err
			}

			err := ReadFromInt32(dst, r.buffer[r.readPosition:], dstLen)
			r.readPosition += dstLen
			if r.lastError != nil {
				r.usageLock.Unlock()
				return available, r.lastError
			}

			r.usageLock.Unlock()
			return dstLen, err
		}

		r.usageLock.Unlock()
		time.Sleep(time.Millisecond * 10)
	}
}

func (r *realtimeStream) SampleRate() int {
	return r.stream.SampleRate()
}
