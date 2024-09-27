package interpreter

import (
	"time"
)

type NativeClockFn struct{}

func (n NativeClockFn) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().UnixMilli()) / 1000.0, nil
}

func (n NativeClockFn) Arity() int {
	return 0
}

func (n NativeClockFn) String() string {
	return "<native fn>"
}
