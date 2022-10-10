package adapter

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/crit/log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func ForEcho(logger *log.Logger) echo.MiddlewareFunc {
	fname := "???"
	lnum := 0

	_, file, line, ok := runtime.Caller(1)

	if ok {
		fname = srcFileParse(file)
		lnum = line
	}

	// format needed by LoggerWithConfig; notice that we are adding data to this string
	// using fmt.Sprintf. Namely, app, src.file and src.line entries. This does have a
	// problem that all logs are going to say the source is the same file/line. But since
	// these logs are usually from the framework itself, its fine for now.
	f := `{"time":"${time_rfc3339}","app":"%s","level":"info","msg":"${method}",` +
		`"data":{"remote":"${remote_ip}","uri":"${uri}","status":${status},` +
		`"latency":"${latency_human}"},"src":{"file":"%s","line":%d}}`

	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: fmt.Sprintf(f, logger.AppName(), fname, lnum),
		Output: logger.Out,
	})
}

// srcFileParse returns either the filename and extension, or the last directory
// (which is also usually the package name in Go) with the filename and extension.
// "project/src/model/user.go" => "model/user.go"
// "main.go" => "main.go
func srcFileParse(filename string) string {
	// "project/src/model/user.go" => "project/src/model", "user.go"
	dir, file := filepath.Split(filename)

	// "project/src/model" => ["project", "src", "model"]
	parts := strings.FieldsFunc(dir, func(c rune) bool {
		return c == filepath.Separator
	})

	if len(parts) > 0 {
		// => "model/user.go"
		return filepath.Join(parts[len(parts)-1], file)
	}

	// "user.go"
	return file
}
