package audio

import (
	"errors"
)

var ErrInvalidReadDestination = errors.New("audio: invalid read destination")

// SliceLength returns the length of a valid audio slice.
func SliceLength(slice interface{}) int {
	switch slice.(type) {
	case []int8:
		return len(slice.([]int8))
	case []int16:
		return len(slice.([]int16))
	case []int32:
		return len(slice.([]int32))
	case []float32:
		return len(slice.([]float32))
	default:
		return 0
	}
}

// ReadFromInt8 converts an []int8 to any other valid audio slice.
func ReadFromInt8(dst interface{}, src []int8, num int) error {
	switch dst.(type) {
	case []int8:
		realDst := dst.([]int8)
		copy(realDst, src)
		return nil
	case []int16:
		realDst := dst.([]int16)
		for i := 0; i < num; i++ {
			realDst[i] = int16(src[i]) << 8
		}
		return nil
	case []int32:
		realDst := dst.([]int32)
		for i := 0; i < num; i++ {
			realDst[i] = int32(src[i]) << 24
		}
		return nil
	case []float32:
		realDst := dst.([]float32)
		for i := 0; i < num; i++ {
			realDst[i] = float32(src[i]) / 128.0
		}
		return nil
	default:
		return ErrInvalidReadDestination
	}
}

// ReadFromInt16 converts an []int16 to any other valid audio slice.
func ReadFromInt16(dst interface{}, src []int16, num int) error {
	switch dst.(type) {
	case []int8:
		realDst := dst.([]int8)
		for i := 0; i < num; i++ {
			realDst[i] = int8(src[i] >> 8)
		}
		return nil
	case []int16:
		realDst := dst.([]int16)
		copy(realDst, src)
		return nil
	case []int32:
		realDst := dst.([]int32)
		for i := 0; i < num; i++ {
			realDst[i] = int32(src[i]) << 16
		}
		return nil
	case []float32:
		realDst := dst.([]float32)
		for i := 0; i < num; i++ {
			realDst[i] = float32(src[i]) / 32768.0
		}
		return nil
	default:
		return ErrInvalidReadDestination
	}
}

// ReadFromInt32 converts an []int32 to any other valid audio slice.
func ReadFromInt32(dst interface{}, src []int32, num int) error {
	switch dst.(type) {
	case []int8:
		realDst := dst.([]int8)
		for i := 0; i < num; i++ {
			realDst[i] = int8(src[i] >> 24)
		}
		return nil
	case []int16:
		realDst := dst.([]int16)
		for i := 0; i < num; i++ {
			realDst[i] = int16(src[i] >> 16)
		}
		return nil
	case []int32:
		realDst := dst.([]int32)
		copy(realDst, src)
		return nil
	case []float32:
		realDst := dst.([]float32)
		for i := 0; i < num; i++ {
			realDst[i] = float32(src[i]) / 128.0
		}
		return nil
	default:
		return ErrInvalidReadDestination
	}
}

// ReadFromFloat32 converts a []float32 to any other valid audio slice.
func ReadFromFloat32(dst interface{}, src []float32, num int) error {
	switch dst.(type) {
	case []int8:
		realDst := dst.([]int8)
		for i := 0; i < num; i++ {
			if src[i] >= 1 {
				realDst[i] = 127
			} else if src[i] <= -1 {
				realDst[i] = -127
			} else {
				realDst[i] = int8(src[i] * 127.0)
			}
		}
		return nil
	case []int16:
		realDst := dst.([]int16)
		for i := 0; i < num; i++ {
			if src[i] >= 1 {
				realDst[i] = 32767
			} else if src[i] <= -1 {
				realDst[i] = -32767
			} else {
				realDst[i] = int16(src[i] * 32767.0)
			}
		}
		return nil
	case []int32:
		realDst := dst.([]int32)
		for i := 0; i < num; i++ {
			if src[i] >= 1 {
				realDst[i] = 2147483647
			} else if src[i] <= -1 {
				realDst[i] = -2147483647
			} else {
				realDst[i] = int32(src[i] * 2147483647.0)
			}
		}
		return nil
	case []float32:
		realDst := dst.([]float32)
		for i := 0; i < num; i++ {
			if src[i] >= 1 {
				realDst[i] = 1
			} else if src[i] <= -1 {
				realDst[i] = -1
			} else {
				realDst[i] = src[i]
			}
		}
		return nil
	default:
		return ErrInvalidReadDestination
	}
}
