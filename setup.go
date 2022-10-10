package log

import (
	"os"
)

// SetupLogger handles the standard logger setup for lambda services.
func SetupLogger(name, build string) *Logger {
	var logLevel = ToLevel(os.Getenv("LOG_LEVEL"))

	return New(name, logLevel).With(Data{"build": build})
}
