package log

import (
	"log"
	"os"
)

// ErrLogger is a logger than write to /dev/stderr
var ErrLogger = log.New(os.Stderr, "", log.LstdFlags)

// OutLogger is a logger than write to /dev/stdout
var OutLogger = log.New(os.Stdout, "", log.LstdFlags)
