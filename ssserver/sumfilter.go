package ssserver

import (
	"math"
)

func (b *Buffer) BeByteafterNew(buf []byte, size ...int) {
	if size != nil {
		for i := 0; i < len(buf); i += 1 {
			b.Buf[b.position] = buf[i]
			b.position++
		}
	} else {
		for i := 0; i < size[0]; i += 1 {
			b.Buf[b.position] = buf[i]
			b.position++
		}
	}
}

func (b *Buffer) Writebytebytype(i interface{}) {
	switch i.(type) {
	case uint8:
		b.Buf[b.position] = byte(i.(uint8))
		b.position++
	case uint16:
		b.Buf[b.position] = byte(i.(uint16))
		b.position++
		b.Buf[b.position] = byte(i.(uint16) >> 8)
		b.position++
	case uint32:
		b.Buf[b.position] = byte(i.(uint32))
		b.position++
		b.Buf[b.position] = byte(i.(uint32) >> 8)
		b.position++
		b.Buf[b.position] = byte(i.(uint32) >> 16)
		b.position++
		b.Buf[b.position] = byte(i.(uint32) >> 24)
		b.position++
	case uint64:
		for n := uint(0); n < uint(Int64); n++ {
			b.Buf[b.position] = byte(i.(uint64) >> (n * 8))
			b.position++
		}
	case int8:
		b.Buf[b.position] = byte(i.(int8))
		b.position++
	case int16:
		b.Buf[b.position] = byte(i.(int16))
		b.position++
		b.Buf[b.position] = byte(i.(int16) >> 8)
		b.position++
	case int32:
		b.Buf[b.position] = byte(i.(int32))
		b.position++
		b.Buf[b.position] = byte(i.(int32) >> 8)
		b.position++
		b.Buf[b.position] = byte(i.(int32) >> 16)
		b.position++
		b.Buf[b.position] = byte(i.(int32) >> 24)
		b.position++
	case int64:
		for n := uint(0); n < uint(Int64); n++ {
			b.Buf[b.position] = byte(i.(int64) >> (n * 8))
			b.position++
		}
	case float32:
		b.Readbytebytype(math.Float32bits(i.(float32)))
	case float64:
		b.Readbytebytype(math.Float64bits(i.(float64)))
	}

}
