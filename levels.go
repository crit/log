package log

import "strings"

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	NoticeLevel
	WarningLevel
	ErrorLevel
	CriticalLevel
	AlertLevel
	EmergencyLevel
)

var defaultLevel = NoticeLevel

var logLevelLabels = map[Level]string{
	DebugLevel:     "debug",
	InfoLevel:      "info",
	NoticeLevel:    "notice",
	WarningLevel:   "warning",
	ErrorLevel:     "error",
	CriticalLevel:  "critical",
	AlertLevel:     "alert",
	EmergencyLevel: "emergency",
}

var logLevelValues = map[string]Level{
	"debug":     DebugLevel,
	"info":      InfoLevel,
	"notice":    NoticeLevel,
	"warning":   WarningLevel,
	"error":     ErrorLevel,
	"critical":  CriticalLevel,
	"alert":     AlertLevel,
	"emergency": EmergencyLevel,
}

func (l Level) String() string {
	return logLevelLabels[l]
}

func ToLevel(value string) Level {
	value = strings.ToLower(value)
	level, ok := logLevelValues[value]

	if !ok {
		return defaultLevel
	}

	return level
}
