package utils

import (
	colorable "github.com/mattn/go-colorable"

	log "github.com/sirupsen/logrus"
)

func LogrusInit(d bool) {
	var debug = d
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
