package exc

import (
	"encoding/gob"
	"fmt"
	"kuroko-linux/util"
	"runtime/debug"
	"strings"
)

type withFields struct {
	fields map[string]any
	msg    string
}

func Fields(args ...any) *withFields {
	return &withFields{fields: util.ArgsToMap(args)}
}

func (wf *withFields) Wrap(e error) error {
	if e == nil {
		return nil
	}
	ret := newKurokoError(e)
	ret.Fields = wf.fields
	ret.Message = wf.msg
	return ret
}

func (wf *withFields) Msg(str string, args ...any) *withFields {
	wf.msg = fmt.Sprintf(str, args...)
	return wf
}

func (wf *withFields) New(str string, args ...any) error {
	return wf.Msg(str, args...).Wrap(nil)
}

func Msg(str string, args ...any) *withFields {
	return Fields().Msg(str, args...)
}

func Wrap(e error) error {
	if e == nil {
		return nil
	}
	if _, ok := e.(*KurokoError); ok {
		return e
	}
	return newKurokoError(e)
}

func New(str string, args ...any) error {
	msg := fmt.Sprintf(str, args...)
	ret := newKurokoError(nil)
	ret.Message = msg
	return ret
}

type KurokoError struct {
	Err     error
	Message string
	Stack   []byte
	Fields  map[string]any
}

func newKurokoError(e error) *KurokoError {
	ret := new(KurokoError)
	ret.Err = e
	var recordStack = true
	if _, ok := e.(*KurokoError); ok {
		recordStack = false
	}
	if recordStack {
		ret.Stack = debug.Stack()
	} else {
		ret.Stack = nil
	}
	return ret
}

func (e *KurokoError) Error() string {
	var err error = e
	builder := strings.Builder{}
	for {
		ke, ok := err.(*KurokoError)
		if ok {
			if len(ke.Message) > 0 {
				builder.WriteString(ke.Message)
				builder.WriteString(": ")
			}
			err = ke.Err
		} else {
			builder.WriteString(err.Error())
			break
		}
	}
	return builder.String()
}

func (e *KurokoError) Unwrap() error { return e.Err }

func init() {
	gob.Register(new(KurokoError))
}
