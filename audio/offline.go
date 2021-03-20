package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
)

// NumberType represents a PCM number type.
type NumberType int

// Possible NumberTypes
const (
	Int8 = iota
	Int16
	Int32
	Float32
)

// ErrInvalidNumberType is returned if an invalid number type is provided,
// mainly to offline streams.
var ErrInvalidNumberType = errors.New("audio: invalid number type")

// OfflineStream represents a stream which is wrapped and is
// usually based on a different stream of bytes or numbers, such as
// over a network.
type OfflineStream struct {
	Stream
	buffer      []int32
	sampleRate  int
	dataChannel chan bool
	closed      bool
	usageLock   *sync.Mutex
	bufferSize  int
}

// NewOfflineStream returns an offline stream, a simple buffered stream that
// allows you to convert io.Read/Write operations into a stream. Buffer size
// is measured by number of samples.
func NewOfflineStream(sampleRate int, bufferSize int) *OfflineStream {
	newStream := &OfflineStream{
		sampleRate:  sampleRate,
		dataChannel: make(chan bool),
		usageLock:   new(sync.Mutex),
		bufferSize:  bufferSize,
	}

	return newStream
}

// ReadBytes reads bytes of values into the offline stream.
func (o *OfflineStream) ReadBytes(rd io.Reader, bo binary.ByteOrder, numType NumberType) error {
	o.usageLock.Lock()
	if o.closed {
		return errors.New("audio: attempt to write to closed OfflineStream")
	}
	o.usageLock.Unlock()

	defer o.Close()

	switch numType {
	case Int8:
		for {
			buffer := make([]byte, o.bufferSize)
			n, err := rd.Read(buffer)
			middle := make([]int8, n)
			for i := 0; i < n; i++ {
				middle[i] = int8(buffer[i])
			}

			if n > 0 {
				convert := make([]int32, n)
				ReadFromInt8(convert, middle, n)

				o.usageLock.Lock()
				o.buffer = append(o.buffer, convert...)
				o.emitDataEvent()
				o.usageLock.Unlock()
			}

			if err != nil {
				return err
			}
		}
	case Int16:
		for {
			buffer := make([]byte, 2*o.bufferSize)
			n, err := rd.Read(buffer)
			middle := make([]int16, n/2)
			for i := 0; i < n/2; i++ {
				middle[i] = int16(bo.Uint16(buffer[i*2:]))
			}

			if n > 0 {
				convert := make([]int32, n)
				ReadFromInt16(convert, middle, n)

				o.usageLock.Lock()
				o.buffer = append(o.buffer, convert...)
				o.emitDataEvent()
				o.usageLock.Unlock()
			}

			if err != nil {
				return err
			}
		}
	case Int32:
		for {
			buffer := make([]byte, 4*o.bufferSize)
			n, err := rd.Read(buffer)
			middle := make([]int32, n/4)
			for i := 0; i < n/4; i++ {
				middle[i] = int32(bo.Uint32(buffer[i*4:]))
			}

			if n > 0 {
				o.usageLock.Lock()
				o.buffer = append(o.buffer, middle...)
				o.emitDataEvent()
				o.usageLock.Unlock()
			}

			if err != nil {
				return err
			}
		}
	case Float32:
		for {
			buffer := make([]byte, 4*o.bufferSize)
			n, err := rd.Read(buffer)
			middle := make([]float32, n/4)
			for i := 0; i < n/4; i++ {
				middle[i] = math.Float32frombits(bo.Uint32(buffer[i*4:]))
			}

			if n > 0 {
				convert := make([]int32, n)
				ReadFromFloat32(convert, middle, n)

				o.usageLock.Lock()
				o.buffer = append(o.buffer, convert...)
				o.emitDataEvent()
				o.usageLock.Unlock()
			}

			if err != nil {
				return err
			}
		}
	default:
		return ErrInvalidNumberType
	}
}

// WriteBytes writes bytes of values into the offline stream.
func (o *OfflineStream) WriteBytes(b []byte, bo binary.ByteOrder, numType NumberType) error {
	rd := bytes.NewReader(b)
	return o.ReadBytes(rd, bo, numType)
}

// WriteValues writes a slice of number values into the offline stream.
func (o *OfflineStream) WriteValues(val interface{}) error {
	o.usageLock.Lock()
	if o.closed {
		return errors.New("audio: attempt to write to closed OfflineStream")
	}
	o.usageLock.Unlock()

	length := SliceLength(val)
	result := make([]int32, length)
	err := ReadFromAnything(result, val, length)
	if err != nil {
		return err
	}

	o.usageLock.Lock()
	o.buffer = append(o.buffer, result...)
	o.emitDataEvent()
	o.usageLock.Unlock()

	return nil
}

// Close closes the offline stream, and causes any ongoing reads to return.
func (o *OfflineStream) Close() {
	o.usageLock.Lock()
	o.closed = true
	o.emitDataEvent()
	o.usageLock.Unlock()
}

func (o *OfflineStream) emitDataEvent() {
	select {
	case o.dataChannel <- true:
	default:
	}
}

func (o *OfflineStream) Read(dst interface{}) (int, error) {
	length := SliceLength(dst)

	for {
		o.usageLock.Lock()

		if o.closed && len(o.buffer) < length {
			if len(o.buffer) == 0 {
				o.usageLock.Unlock()
				return 0, io.EOF
			}

			last := len(o.buffer)
			ReadFromInt32(dst, o.buffer, last)
			o.buffer = nil
			o.usageLock.Unlock()
			return last, nil
		}

		if len(o.buffer) >= length {
			ReadFromInt32(dst, o.buffer, length)
			o.buffer = o.buffer[length:]
			o.usageLock.Unlock()
			return length, nil
		}
		o.usageLock.Unlock()

		<-o.dataChannel
	}
}

// SampleRate returns the sample rate of the offline stream.
func (o *OfflineStream) SampleRate() int {
	return o.sampleRate
}
