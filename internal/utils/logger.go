package utils

import "log"

func Info(format string, args ...any) {
	log.Printf("[info] "+format, args...)
}

func Error(format string, args ...any) {
	log.Printf("[error] "+format, args...)
}
