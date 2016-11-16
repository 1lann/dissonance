package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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
}

// NewOfflineStream returns an offline stream, a simple buffered stream that
// allows you to convert io.Read/Write operations into a stream.
func NewOfflineStream(sampleRate int) *OfflineStream {
	newStream := &OfflineStream{
		sampleRate:  sampleRate,
		dataChannel: make(chan bool),
		usageLock:   new(sync.Mutex),
	}

	return newStream
}

// ReadBytes reads bytes of values into the offline stream.
func (o *OfflineStream) ReadBytes(rd io.Reader, bo binary.ByteOrder, numType NumberType) error {
	for {
		switch numType {
		case Int8:
			var num int8
			err := binary.Read(rd, bo, &num)
			if err != nil {
				return err
			}

			buffer := []int32{0}
			ReadFromInt8(buffer, []int8{num}, 1)

			o.usageLock.Lock()
			o.buffer = append(o.buffer, buffer...)
			o.emitDataEvent()
			o.usageLock.Unlock()
		case Int16:
			var num int16
			err := binary.Read(rd, bo, &num)
			if err != nil {
				return err
			}

			buffer := []int32{0}
			ReadFromInt16(buffer, []int16{num}, 1)

			o.usageLock.Lock()
			o.buffer = append(o.buffer, buffer...)
			o.emitDataEvent()
			o.usageLock.Unlock()
		case Int32:
			var num int32
			err := binary.Read(rd, bo, &num)
			if err != nil {
				return err
			}

			buffer := []int32{0}
			ReadFromInt32(buffer, []int32{num}, 1)

			o.usageLock.Lock()
			o.buffer = append(o.buffer, buffer...)
			o.emitDataEvent()
			o.usageLock.Unlock()
		case Float32:
			var num float32
			err := binary.Read(rd, bo, &num)
			if err != nil {
				return err
			}

			buffer := []int32{0}
			ReadFromFloat32(buffer, []float32{num}, 1)

			o.usageLock.Lock()
			o.buffer = append(o.buffer, buffer...)
			o.emitDataEvent()
			o.usageLock.Unlock()
		default:
			return ErrInvalidNumberType
		}
	}
}

// WriteBytes writes bytes of values into the offline stream.
func (o *OfflineStream) WriteBytes(b []byte, bo binary.ByteOrder, numType NumberType) error {
	rd := bytes.NewReader(b)
	return o.ReadBytes(rd, bo, numType)
}

// WriteValues writes a slice of number values into the offline stream.
func (o *OfflineStream) WriteValues(val interface{}) error {
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

		if o.closed {
			o.usageLock.Unlock()
			return 0, io.EOF
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
