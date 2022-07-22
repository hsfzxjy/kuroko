package log

import (
	"fmt"
	"io"
	"kuroko-linux/internal/exc"
	"kuroko-linux/util"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	TRACE LogLevel = (1 + iota) * 10
	DEBUG
	INFO
	WARNING
	ERROR
)

var logLevelToShortName = map[LogLevel]string{
	TRACE:   "TRCE",
	DEBUG:   "DEBG",
	INFO:    "INFO",
	WARNING: "WARN",
	ERROR:   "ERRO",
}

var output io.Writer = os.Stderr
var lock = sync.Mutex{}

func SetOutput(out io.Writer) {
	output = out
}

type Logger struct {
	parent *Logger
	name   string
	fields map[string]any
}

type hasName interface{ Name() string }

func For[T hasName](val T) *Logger {
	return Name(val.Name())
}

func Name(name string) *Logger {
	return &Logger{
		parent: nil,
		name:   name,
		fields: nil,
	}
}

func (l *Logger) Fields(args ...any) *Logger {
	return &Logger{
		parent: l,
		name:   l.name,
		fields: util.ArgsToMap(args),
	}
}

func (l *Logger) Clone() *Logger {
	return &Logger{
		parent: l,
		name:   l.name,
		fields: nil,
	}
}

func (l *Logger) AddFields(args ...any) {
	l.fields = util.MergeMap(l.fields, util.ArgsToMap(args))
}

func (l *Logger) Trace(msg string, args ...any) {
	l.log(TRACE, msg, args)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(DEBUG, msg, args)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(INFO, msg, args)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(WARNING, msg, args)
}

func (l *Logger) Errorf(e any, msg string, args ...any) {
	l.logError(ERROR, e, msg, args)
}

func (l *Logger) Error(e any) {
	l.logError(ERROR, e, "Caught error", nil)
}

func (l *Logger) Panic(e any) {
	l.logError(ERROR, e, "Caught panic", nil)
}

func maxKeyLen(m map[string]any, basis int) int {
	maxlen := basis
	for key := range m {
		length := len(key)
		if length > maxlen {
			maxlen = length
		}
	}
	return maxlen
}

func (l *Logger) maxKeyLen(basis int) int {
	var maxlen = basis
	for l != nil {
		maxlen = maxKeyLen(l.fields, maxlen)
		l = l.parent
	}
	return maxlen
}

func (lr *Logger) builder(level LogLevel, msg string, args []any) *strings.Builder {
	builder := new(strings.Builder)
	fmt.Fprintf(builder, "[%d] [%s] ", os.Getpid(), logLevelToShortName[level])
	builder.WriteString(time.Now().Format("2006-01-02 15:04:05.000"))
	fmt.Fprintf(builder, " [%s]", lr.name)
	builder.WriteRune(' ')
	msg = strings.TrimRight(util.FormatList(msg, args), "\r\n")
	builder.WriteString(msg)
	builder.WriteRune('\n')
	spec := fmt.Sprintf(" %%-%ds = %%+v\n", lr.maxKeyLen(0))

	for lr != nil {
		for key, value := range lr.fields {
			fmt.Fprintf(builder, spec, key, value)
		}
		lr = lr.parent
	}
	return builder
}

func (lr *Logger) log(level LogLevel, msg string, args []any) {
	builder := lr.builder(level, msg, args)
	lock.Lock()
	output.Write([]byte(builder.String()))
	lock.Unlock()
}

func getErrMsg(e any) string {
	var errMsg string
	if ke, ok := e.(*exc.KurokoError); ok {
		errMsg = ke.Message
	} else {
		errMsg = fmt.Sprintf("%+v", e)
	}
	return errMsg
}

func (lr *Logger) logError(level LogLevel, e any, msg string, args []any) {
	lr.fields = util.MergeMap(lr.fields, map[string]any{
		"Error": getErrMsg(e),
	})
	builder := lr.builder(level, msg, args)

	indent := "  "
	for {
		ke, ok := e.(*exc.KurokoError)
		if !ok {
			break
		}
		spec := fmt.Sprintf(
			"%s%%-%ds = %%+v\n",
			indent, maxKeyLen(ke.Fields, 7),
		)
		for key, value := range ke.Fields {
			fmt.Fprintf(builder, spec, key, value)
		}
		wrapped := ke.Unwrap()
		fmt.Fprintf(
			builder, spec,
			"Wrapped", getErrMsg(wrapped),
		)
		if ke.Stack != nil {
			builder.WriteString(indent + "STACK: ")
			replacer := strings.NewReplacer(
				"\n", "\n"+indent,
				"\t", " ",
			)
			replacer.WriteString(builder, string(ke.Stack))
		}
		indent += " "
		e = wrapped
	}
	lock.Lock()
	output.Write([]byte(builder.String()))
	lock.Unlock()
}
