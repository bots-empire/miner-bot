package log

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	tagOk    = color.HiGreenString("OK ")
	tagInfo  = color.HiBlueString("INF")
	tagWarn  = color.HiYellowString("WRN")
	tagError = color.HiRedString("ERR")
)

func createTag(tag, prefix string) string {
	return fmt.Sprintf("[%s] %s ", tag, color.HiWhiteString(prefix+":"))
}

type Logger interface {
	Prefix(prefix string) Logger

	Ok(format string, values ...interface{})
	Info(format string, values ...interface{})
	Warn(format string, values ...interface{})
	Fatal(format string, values ...interface{})
}
