package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// Extracts the project root directory from the current file path
func getProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root without finding go.mod, return current dir as fallback
			return filepath.Dir(file)
		}
		dir = parentDir
	}
}

// sourceRelativeToRoot converts absolute source path to relative from project root
func sourceRelativeToRoot(projectRoot string) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				relPath, err := filepath.Rel(projectRoot, source.File)
				if err == nil {
					source.File = relPath
				}
			}
		}
		return a
	}
}

// SetupLogger configures and returns a logger based on the environment
func SetupLogger(env string) *slog.Logger {
	projectRoot := getProjectRoot()

	var handler slog.Handler

	switch env {
	case EnvLocal:
		// Local: Text format, Debug level, with source and time
		opts := &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: sourceRelativeToRoot(projectRoot),
		}
		handler = slog.NewTextHandler(os.Stdout, opts)

	case EnvDev:
		// Dev: JSON format, Error level, with additional debugging fields
		opts := &slog.HandlerOptions{
			Level:     slog.LevelError,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// First handle source path conversion
				if a.Key == slog.SourceKey {
					a = sourceRelativeToRoot(projectRoot)(groups, a)
				}

				// Then handle other attributes
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   a.Key,
						Value: a.Value,
					}
				}
				return a
			},
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)

	case EnvProd:
		// Prod: JSON format, Info level, with structured output
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Remove source information in production
				if a.Key == slog.SourceKey {
					return slog.Attr{}
				}
				return a
			},
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)

	default:
		// Fallback to basic logger with warning level
		opts := &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
