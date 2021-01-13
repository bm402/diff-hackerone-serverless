package main

import (
	"errors"
	"log"
)

// Writes logs and error logs; on error, sends a Slack notification and exits the program
func logger(message interface{}) {
	switch message.(type) {
	case string:
		log.Print(message)
	case error:
		sendSlackErrorNotification(message.(error))
		log.Fatal(message)
	default:
		logger(errors.New("Unknown log type"))
	}
}
