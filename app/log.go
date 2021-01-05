package main

import (
	"errors"
	"log"
)

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
