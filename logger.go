package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	level Level
	app   string
	data  Data
	mutex sync.Mutex
	out   io.Writer
}

type Loggable interface {
	Log() map[string]any
}

type Data map[string]any

func (d Data) Log() map[string]any {
	return d
}

// New creates a new Logger instance with a specific name and the minimum log level to write.
func New(app string, logLevel Level) *Logger {
	return &Logger{
		level: logLevel,
		app:   app,
		out:   &stdOutWriter{},
	}
}

// Debug records detailed debug information about the data.
func (l *Logger) Debug(msg string, args ...any) {
	l.output(2, DebugLevel, fmt.Sprintf(msg, args...))
}

// Info records interesting events. Examples: User logs in, SQL logs, etc.
func (l *Logger) Info(msg string, args ...any) {
	l.output(2, InfoLevel, fmt.Sprintf(msg, args...))
}

// Notice records normal but significant events.
func (l *Logger) Notice(msg string, args ...any) {
	l.output(2, NoticeLevel, fmt.Sprintf(msg, args...))
}

// Warn records exceptional occurrences that are not errors. Examples: Use of deprecated APIs,
// poor use of an API, undesirable things that are not necessarily wrong.
func (l *Logger) Warn(msg string, args ...any) {
	l.output(2, WarningLevel, fmt.Sprintf(msg, args...))
}

// Error records runtime errors that do not require immediate action but should typically be logged and monitored.
func (l *Logger) Error(msg string, args ...any) {
	l.output(2, ErrorLevel, fmt.Sprintf(msg, args...))
}

// Critical records critical conditions. Example: Application service unavailable, unexpected exception.
func (l *Logger) Critical(msg string, args ...any) {
	l.output(2, CriticalLevel, fmt.Sprintf(msg, args...))
}

// Alert records exceptions where action MUST be taken immediately. Example: website down, database unavailable, etc.
// This should wake someone up.
func (l *Logger) Alert(msg string, args ...any) {
	l.output(2, AlertLevel, fmt.Sprintf(msg, args...))
}

// Emergency records instances where the system is totally unusable.
func (l *Logger) Emergency(msg string, args ...any) {
	l.output(2, EmergencyLevel, fmt.Sprintf(msg, args...))
}

// Fatal writes and emergency log and then calls os.Exit(1).
func (l *Logger) Fatal(msg string, args ...any) {
	l.output(2, EmergencyLevel, fmt.Sprintf(msg, args...))
	os.Exit(1)
}

// With saves specific data to be written out to the remote service when the level is called.
func (l *Logger) With(data ...Loggable) *Logger {
	// set will allow us to detect when a key has already been created with a value
	// and change the value to a slice of values if the key is presented again for logging.
	// logger := log.New(...).With(log.Data{"key": "v1"}) => {... "key":"v1" ...}
	// logger.With(log.Data{"key": "v2"})                 => {... "key":["v1","v2"] ...}
	set := make(map[string]any)

	// @IMPROVE Right now this implementation is memory hungry. There is much room for improvement.
	for key, value := range l.data {
		set[key] = value
	}

	for _, node := range data {
		for key, value := range node.Log() {
			// do we have a current key already?
			if current, ok := set[key]; ok {
				// is the current value already a slice?
				if s, ok := current.([]any); ok {
					// append to old slice
					s = append(s, value)
					set[key] = s
					continue
				}

				// create a new slice since we have the key already but a new value
				set[key] = []any{current, value}
				continue
			}

			// create a new entry
			set[key] = value
		}
	}

	return &Logger{
		app:   l.app,
		level: l.level,
		data:  set,
		out:   l.out,
	}
}

// output creates the structured log and sends it to the writer.
func (l *Logger) output(callDepth int, level Level, msg string) {
	var out WriteLog
	var ok bool

	if level < l.level {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.data = make(map[string]any)
		return
	}

	out.Time = time.Now().UTC()

	_, out.Src.File, out.Src.Line, ok = runtime.Caller(callDepth)

	if !ok {
		out.Src.File = "???"
		out.Src.Line = 0
	} else {
		out.Src.TruncateFile()
	}

	out.Level = level.String()
	out.Msg = msg
	out.Data = map[string]any{}

	l.mutex.Lock()

	out.App = l.app

	for key, value := range l.data {
		out.Data[key] = value
	}

	l.data = make(map[string]any)

	l.mutex.Unlock()

	data, err := json.Marshal(out)

	if err != nil {
		data = []byte("Logger unable to marshal log output to JSON: " + err.Error())
	}

	_, _ = l.out.Write(data)
}

type WriteLog struct {
	Time  time.Time `json:"time"`
	App   string    `json:"app"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Data  Data      `json:"data,omitempty"`
	Src   Src       `json:"Src"`
}

type Src struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// TruncateFile mutates the file string into either the filename and extension,
// or the last directory (which is also usually the package name in Go) with the filename
// and extension.
//
// "project/Src/model/user.go" => "model/user.go"
// "main.go" => "main.go"
func (s *Src) TruncateFile() {
	// "project/Src/model/user.go" => "project/Src/model", "user.go"
	dir, file := filepath.Split(s.File)

	// "project/Src/model" => ["project", "Src", "model"]
	parts := strings.FieldsFunc(dir, func(r rune) bool {
		return r == filepath.Separator
	})

	if len(parts) > 0 {
		// => "model/user.go"
		s.File = filepath.Join(parts[len(parts)-1], file)
	} else {
		s.File = file // "user.go"
	}
}
