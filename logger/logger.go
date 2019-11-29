package logger

import (
	"fmt"
	"log"
	"os"
)

var w bool

func SetLogger(wError bool) {
	w = wError
}

func Warning(err error, msg string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if w {
		log.Fatal(msg)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}
