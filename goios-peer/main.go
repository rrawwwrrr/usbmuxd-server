package main

import (
	"goios-peer/goios"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.TraceLevel)
	goios.Start()
}
