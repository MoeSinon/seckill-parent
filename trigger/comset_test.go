package trigger

import (
	"encoding/json"
	"testing"
	"unsafe"

	"go.uber.org/zap"
)

type sstableinfo struct {
	ver         int
	Postart     int
	length      int
	indexstart  int
	indexlength int
	part        int
	Rw          string
	Log         *zap.SugaredLogger
}

func makestruct() sstableinfo {
	return sstableinfo{
		ver:         1,
		Postart:     0,
		length:      0,
		indexstart:  0,
		indexlength: 0,
		part:        0,
		Rw:          "",
	}
}

func jaso() ([]byte, error) {
	return json.Marshal(makestruct())

}

func TestXxx(t *testing.T) {
	butes, _ := jaso()
	_ = json.Unmarshal(butes, &sstableinfo{})
	t.Fatal(*(*string)(unsafe.Pointer(&butes)))
}
