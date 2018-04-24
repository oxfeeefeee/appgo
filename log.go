package appgo

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	// "gopkg.in/gemnasium/logrus-graylog-hook.v1"
	"gitlab.wallstcn.com/wscnbackend/ivankastd"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

var once sync.Once

func initLog() {
	// logrus.SetLevel(Conf.LogLevel)
	ivankastd.InitLog(Conf.IvankaLog)
}

func InitLogHooks(skip int) {
	_, file, _, _ := runtime.Caller(skip)
	once.Do(func() {
		rootDir := filepath.Dir(file)
		fiHook := &FileInfoHook{rootDir}
		logrus.AddHook(fiHook)
		// gladdr := Conf.Graylog.Ip + ":" + Conf.Graylog.Port
		// glhook := graylog.NewGraylogHook(gladdr, Conf.Graylog.Facility, nil)
		// logrus.AddHook(glhook)
	})
}

func SetLogFile(f *os.File, withstd bool) {
	syscall.Dup2(int(f.Fd()), 2)
	if withstd {
		logrus.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		logrus.SetOutput(f)
	}
}

type FileInfoHook struct {
	rootDir string
}

func (hook *FileInfoHook) Fire(entry *logrus.Entry) error {
	entry.Data["file"] = fileInfo(hook.rootDir, 8)
	return nil
}

func (hook *FileInfoHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func fileInfo(rootDir string, skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 0
	} else {
		if rel, err := filepath.Rel(rootDir, file); err != nil {
			file = "<???>"
		} else {
			file = rel
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}
