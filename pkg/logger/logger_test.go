package logger

import (
	"os"
	"testing"
)

func Test_Setup(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Nominal case",
			args: args{
				cfg: &Config{
					Level:      "info",
					Format:     "json",
					Graylog:    GraylogConfig{Enabled: false},
					EnableFile: false,
				},
			},
		},
		{
			name: "Error case - invalid log level",
			args: args{
				cfg: &Config{
					Level:      "invalid",
					Format:     "text",
					Graylog:    GraylogConfig{Enabled: false},
					EnableFile: false,
				},
			},
		},
		{
			name: "Error case - file logging enabled but no file path",
			args: args{
				cfg: &Config{
					Level:      "info",
					Format:     "text",
					Graylog:    GraylogConfig{Enabled: false},
					EnableFile: true,
					FilePath:   "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(tt.args.cfg)
		})
	}
}

func Test_addGraylogHook(t *testing.T) {
	type args struct {
		cfg GraylogConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Nominal case",
			args: args{
				cfg: GraylogConfig{
					Enabled:  true,
					URL:      "http://localhost",
					Port:     12201,
					Facility: "test-facility",
				},
			},
		},
		{
			name: "Error case - Graylog disabled",
			args: args{
				cfg: GraylogConfig{
					Enabled:  false,
					URL:      "http://localhost",
					Port:     12201,
					Facility: "test-facility",
				},
			},
		},
		{
			name: "Error case - Invalid URL",
			args: args{
				cfg: GraylogConfig{
					Enabled:  true,
					URL:      "",
					Port:     12201,
					Facility: "test-facility",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addGraylogHook(tt.args.cfg)
		})
	}
}

func Test_getHostname(t *testing.T) {
	t.Run("nominal", func(t *testing.T) {
		want := func() string {
			hostname, _ := os.Hostname()
			return hostname
		}()
		if got := getHostname(); got != want {
			t.Errorf("getHostname() = %v, want %v", got, want)
		}
	})
}
