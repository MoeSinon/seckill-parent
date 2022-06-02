package ssserver

import (
	"errors"
	"io"
)

const (
	// See http://golang.org/ref/spec#Numeric_types

	// SizeUint8 is the byte size of a uint8.
	Uint8 = 1
	// SizeUint16 is the byte size of a uint16.
	Uint16 = 2
	// SizeUint32 is the byte size of a uint32.
	Uint32 = 4
	// SizeUint64 is the byte size of a uint64.
	Uint64 = 8

	// SizeInt8 is the byte size of a int8.
	Int8 = 1
	// SizeInt16 is the byte size of a int16.
	Int16 = 2
	// SizeInt32 is the byte size of a int32.
	Int32 = 4
	// SizeInt64 is the byte size of a int64.
	Int64 = 8

	// SizeFloat32 is the byte size of a float32.
	Float32 = 4
	// SizeFloat64 is the byte size of a float64.
	Float64 = 8

	// SizeByte is the byte size of a byte.
	// The `byte` type is aliased (by Go definition) to uint8.
	Byte = 1

	// SizeBool is the byte size of a bool.
	// The `bool` type is aliased (by flatbuffers convention) to uint8.
	Bool = 1

	// SizeSOffsetT is the byte size of an SOffsetT.
	// The `SOffsetT` type is aliased (by flatbuffers convention) to int32.
	SOffsetT = 4
	// SizeUOffsetT is the byte size of an UOffsetT.
	// The `UOffsetT` type is aliased (by flatbuffers convention) to uint32.
	UOffsetT = 4
	// SizeVOffsetT is the byte size of an VOffsetT.
	// The `VOffsetT` type is aliased (by flatbuffers convention) to uint16.
	VOffsetT = 2
)

type Buffer struct {
	Buf      []byte
	position int
}

func NewBuffer(buf []byte, size ...int) *Buffer {
	b := &Buffer{}
	if size != nil {
		if len(buf) == size[0] {
			b.Buf = buf
			b.position = 0
			return b
		} else {
			b.Buf = make([]byte, size[0])
			b.position = 0
			copy(b.Buf, buf)
			return b
		}

	} else {
		b.Buf = buf
		b.position = 0
		return b
	}
	return b
}

func (b *Buffer) Readbytebytype(i interface{}) (interface{}, error) {
	// if reflect.TypeOf(i).Kind() != reflect.Ptr{
	// 	if reflect.TypeOf(i).Elem().Kind() != reflect.Struct{
	// 		return nil,errors.New("not a struct and not a pointer")
	// 	}
	// }
	// types := reflect.TypeOf(i).Elem().Kind()
	// switch types {
	// case reflect.Uint8:
	// 	buf, err := b.GetBytes(SizeUint8)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	return uint8(buf[0]), nil
	// case reflect.Uint16:
	// 	var n uint16
	// 	buf, err := b.GetBytes(SizeUint16)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= uint16(buf[0])
	// 	n |= uint16(buf[1]) << 8
	// 	return n, nil
	// case reflect.Uint32:
	// 	var n uint32
	// 	buf, err := b.GetBytes(SizeUint32)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= uint32(buf[0])
	// 	n |= uint32(buf[1]) << 8
	// 	n |= uint32(buf[2]) << 16
	// 	n |= uint32(buf[3]) << 24
	// 	return n, nil
	// case reflect.Uint64:
	// 	var n uint64
	// 	buf, err := b.GetBytes(SizeUint64)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= uint64(buf[0])
	// 	n |= uint64(buf[1]) << 8
	// 	n |= uint64(buf[2]) << 16
	// 	n |= uint64(buf[3]) << 24
	// 	n |= uint64(buf[4]) << 32
	// 	n |= uint64(buf[5]) << 40
	// 	n |= uint64(buf[6]) << 48
	// 	n |= uint64(buf[7]) << 56
	// 	return n, nil
	// case reflect.Int8:
	// 	buf, err := b.GetBytes(SizeInt8)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	return int8(buf[0]), nil
	// case reflect.Int16:
	// 	var n int16
	// 	buf, err := b.GetBytes(SizeInt16)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= int16(buf[0])
	// 	n |= int16(buf[1]) << 8
	// 	return n, nil
	// case reflect.Int32:
	// 	var n int32
	// 	buf, err := b.GetBytes(SizeInt32)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= int32(buf[0])
	// 	n |= int32(buf[1]) << 8
	// 	n |= int32(buf[2]) << 16
	// 	n |= int32(buf[3]) << 24
	// 	return n, nil
	// case reflect.Int64:
	// 	var n int64
	// 	buf, err := b.GetBytes(SizeInt64)
	// 	if err != nil {
	// 		return 0, nil
	// 	}
	// 	n |= int64(buf[0])
	// 	n |= int64(buf[1]) << 8
	// 	n |= int64(buf[2]) << 16
	// 	n |= int64(buf[3]) << 24
	// 	n |= int64(buf[4]) << 32
	// 	n |= int64(buf[5]) << 40
	// 	n |= int64(buf[6]) << 48
	// 	n |= int64(buf[7]) << 56
	// 	return n, nil
	// }

	switch i.(type) {
	case uint8:
		buf, err := b.GetBytes(Uint8)
		if err != nil {
			return nil, nil
		}
		return uint8(buf[0]), nil
	case uint16:
		var n uint16
		buf, err := b.GetBytes(Uint16)
		if err != nil {
			return nil, nil
		}
		n |= uint16(buf[0])
		n |= uint16(buf[1]) << 8
		return n, nil
	case uint32:
		var n uint32
		buf, err := b.GetBytes(Uint32)
		if err != nil {
			return nil, nil
		}
		n |= uint32(buf[0])
		n |= uint32(buf[1]) << 8
		n |= uint32(buf[2]) << 16
		n |= uint32(buf[3]) << 24
		return n, nil
	case uint64:
		var n uint64
		buf, err := b.GetBytes(Uint64)
		if err != nil {
			return nil, nil
		}
		for n := uint64(0); n < uint64(Uint64); n++ {
			n |= uint64(buf[n]) << (n * 8)
		}
		return n, nil
	case int8:
		buf, err := b.GetBytes(Int8)
		if err != nil {
			return nil, nil
		}
		return int8(buf[0]), nil
	case int16:
		var n int16
		buf, err := b.GetBytes(Int16)
		if err != nil {
			return nil, nil
		}
		n |= int16(buf[0])
		n |= int16(buf[1]) << 8
		return n, nil
	case int32:
		var n int32
		buf, err := b.GetBytes(Int32)
		if err != nil {
			return nil, nil
		}
		n |= int32(buf[0])
		n |= int32(buf[1]) << 8
		n |= int32(buf[2]) << 16
		n |= int32(buf[3]) << 24
		return n, nil
	case int64:
		var n int64
		buf, err := b.GetBytes(Int64)
		if err != nil {
			return nil, nil
		}
		for n := int64(0); n < int64(Int64); n++ {
			n |= int64(buf[n]) << (n * 8)
		}
		return n, nil
	}
	return nil, errors.New("parsed wrong type")

}

func (b *Buffer) GetBytes(length int) ([]byte, error) {
	bufferLength := len(b.Buf)
	bufferWindow := b.position + length
	if bufferLength < length {
		return nil, io.EOF
	}
	if bufferWindow > bufferLength {
		return nil, io.EOF
	}
	value := b.Buf[b.position:bufferWindow]
	b.position += length
	return value, nil
}
