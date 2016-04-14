// +build !windows

// Copyright 2014 AirilyApp

package log

import (
	"os"
	"syscall"
)

func setCrashLogFile(f *os.File) {
	syscall.Dup2(int(f.Fd()), 2)
}
