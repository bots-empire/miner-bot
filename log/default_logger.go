package log

import (
	"fmt"
	stdlog "log"
	"os"
)

type defaultLogger struct {
	file   os.File
	inner  *stdlog.Logger
	prefix string
}

func NewDefaultLogger() Logger {
	flags := stdlog.Ldate | stdlog.Lmicroseconds
	return &defaultLogger{
		inner:  stdlog.New(os.Stdout, "", flags),
		prefix: "",
	}
}

func (l *defaultLogger) Prefix(prefix string) Logger {
	return &defaultLogger{
		inner:  l.inner,
		prefix: prefix,
	}
}

func (l *defaultLogger) Ok(format string, values ...interface{}) {
	l.inner.Print(createTag(tagOk, l.prefix) + fmt.Sprintf(format, values...))
}

func (l *defaultLogger) Info(format string, values ...interface{}) {
	l.inner.Print(createTag(tagInfo, l.prefix) + fmt.Sprintf(format, values...))
}

func (l *defaultLogger) Warn(format string, values ...interface{}) {
	l.inner.Print(createTag(tagWarn, l.prefix) + fmt.Sprintf(format, values...))
}

func (l *defaultLogger) Fatal(format string, values ...interface{}) {
	l.inner.Fatal(createTag(tagError, l.prefix) + fmt.Sprintf(format, values...))
}
