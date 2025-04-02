package factory

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	databasesql "github.com/tezos-delegation-service/internal/adapter/database/impl/psql"
	metricsnoop "github.com/tezos-delegation-service/internal/adapter/metrics/impl/noop"
)

func Test_Implementation_String(t *testing.T) {
	tests := []struct {
		name string
		i    Implementation
		want string
	}{
		{
			name: "Nominal case",
			i:    ImplPSQL,
			want: "psql",
		},
		{
			name: "Error case - empty implementation",
			i:    Implementation(""),
			want: "",
		},
		{
			name: "Error case - unknown implementation",
			i:    Implementation("unknown"),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_New(t *testing.T) {
	os.Setenv("GO_TESTING", "1")
	defer os.Unsetenv("GO_TESTING")

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Nominal case - PSQL",
			cfg: Config{
				Impl: ImplPSQL,
				psql: &databasesql.Config{
					Host:     "localhost",
					Port:     5432,
					User:     "user",
					Password: "password",
					DBName:   "tezos",
					SSLMode:  "disable",
				},
			},
			wantErr: false,
		},
		{
			name: "Error case - Missing PSQL config",
			cfg: Config{
				Impl: ImplPSQL,
				psql: nil,
			},
			wantErr: true,
		},
		{
			name: "Error case - Unsupported implementation",
			cfg: Config{
				Impl: Implementation("unsupported"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.cfg, metricsnoop.New())
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.NotNil(t, got, "Expected non-nil adapter when no error")
			} else {
				assert.Nil(t, got, "Expected nil adapter when error occurs")
			}
		})
	}
}
