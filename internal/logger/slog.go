package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"doozip/internal/utils"
)

const (
	EnvDev  = "development"
	EnvProd = "production"
)

// sourceRelativeToRoot converts absolute source path to relative from project root
func SourceRelativeToRoot(projectRoot string) func(groups []string, a slog.Attr) slog.Attr {
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
	projectRoot := utils.GetProjectRoot()

	var handler slog.Handler

	switch env {
	case EnvDev:
		// Local: Text format, Debug level, with source and time
		opts := &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: SourceRelativeToRoot(projectRoot),
		}
		handler = slog.NewTextHandler(os.Stdout, opts)

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
