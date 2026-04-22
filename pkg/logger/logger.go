package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type structuredLogger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

var (
	instance *structuredLogger
	once     sync.Once
)

func Init() {
	once.Do(func() {
		instance = &structuredLogger{
			infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
			warnLogger:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
			errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		}
	})
}

func Info(message string, keyvals ...any) {
	if instance == nil {
		Init()
	}
	logWith(instance.infoLogger, message, keyvals...)
}

func Warn(message string, keyvals ...any) {
	if instance == nil {
		Init()
	}
	logWith(instance.warnLogger, message, keyvals...)
}

func Error(message string, keyvals ...any) {
	if instance == nil {
		Init()
	}
	logWith(instance.errorLogger, message, keyvals...)
}

func Fatal(message string, keyvals ...any) {
	if instance == nil {
		Init()
	}
	logWith(instance.errorLogger, message, keyvals...)
}

func logWith(target *log.Logger, message string, keyvals ...any) {
	if instance == nil || target == nil {
		Init()
		target = instance.infoLogger
	}

	parts := []string{message}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 >= len(keyvals) {
			parts = append(parts, fmt.Sprintf("%v", keyvals[i]))
			break
		}
		parts = append(parts, fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1]))
	}

	target.Println(strings.Join(parts, " "))
}
