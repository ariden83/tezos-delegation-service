package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

// Log is the global logger instance
var Log *logrus.Logger

// Config represents the logging configuration.
type Config struct {
	Level      string
	Format     string
	Graylog    GraylogConfig
	EnableFile bool   `mapstructure:"enable_file"`
	FilePath   string `mapstructure:"file_path"`
}

// GraylogConfig represents the Graylog configuration.
type GraylogConfig struct {
	Enabled  bool
	URL      string
	Port     int
	Facility string
}

// Setup configures the logger based on the provided configuration.
func Setup(cfg *Config) {
	Log = logrus.New()

	switch cfg.Level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	if cfg.Format == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	Log.Out = os.Stdout

	if cfg.EnableFile && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Log.Out = file
		} else {
			Log.Warnf("Failed to log to file %s: %v", cfg.FilePath, err)
		}
	}

	if cfg.Graylog.Enabled && cfg.Graylog.URL != "" {
		addGraylogHook(cfg.Graylog)
	}

	Log.Info("Logger initialized")
}

// addGraylogHook adds a Graylog hook to the logger.
func addGraylogHook(cfg GraylogConfig) {
	graylogURL := cfg.URL
	if cfg.Port > 0 {
		graylogURL = graylogURL + ":" + string(rune(cfg.Port))
	}

	hook := graylog.NewGraylogHook(graylogURL, map[string]interface{}{
		"facility": cfg.Facility,
		"host":     getHostname(),
	})

	Log.AddHook(hook)
	Log.Info("Graylog logging enabled")
}

// getHostname returns the hostname of the current machine.
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
