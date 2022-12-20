//go:build !release
// +build !release

package mytool

import (
	"fmt"
	"io"
	"log"
	"os"
)

func Log(msg ...interface{}) {
	log.Output(2, fmt.Sprintln(msg...))
}

func Logf(format string, msg ...interface{}) {
	log.Output(2, fmt.Sprintf(format, msg...))
}

func GetLogWriter() io.Writer {
	return os.Stdout
}

func CloseLogWriter() {

}
