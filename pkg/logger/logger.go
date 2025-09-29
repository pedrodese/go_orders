package logger

import (
	"log"
	"os"
)

// Simple logger setup para começar
func Init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[ORDER-SERVICE] ")
}

func Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	log.Fatalf("FATAL: "+msg, args...)
	os.Exit(1)
}

func Debug(msg string, args ...interface{}) {
	if os.Getenv("ENV") != "production" {
		log.Printf("DEBUG: "+msg, args...)
	}
}
