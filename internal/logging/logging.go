package logging

import (
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"
)

// NewLogger provides a new logger based on the environment type
func NewLogger() *zap.SugaredLogger {
	dev, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	var l *zap.Logger
	var err error

	// TODO: Figure out some sane options here

	if dev {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		// Just blow up for now
		log.Fatalf("error creating logger: %s", err)
	}

	return l.Sugar()
}
