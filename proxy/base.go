package proxy

import "bytes"

type DataType int

const (
	GDJ0892018A DataType = iota
	GDJ0892018B
	GDJ0892018C
	GDJ0892018D
	GDJ0892018E
)

type DataPack struct {
	Type   DataType
	Buffer *bytes.Buffer
}
