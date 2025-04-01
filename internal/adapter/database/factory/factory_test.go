package factory

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/adapter/database"
	databasemock "github.com/tezos-delegation-service/internal/adapter/database/impl/mock"
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
	tests := []struct {
		name    string
		cfg     Config
		want    database.Adapter
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
			want: func() database.Adapter {
				db, err := databasesql.New(databasesql.Config{
					Host:     "localhost",
					Port:     5432,
					User:     "user",
					Password: "password",
					DBName:   "tezos",
					SSLMode:  "disable",
				})
				assert.NoError(t, err)
				assert.NotNil(t, db)
				return db
			}(),
			wantErr: false,
		},
		{
			name: "Error case - Missing PSQL config",

			cfg: Config{
				Impl: ImplPSQL,
				psql: nil,
			},

			want:    nil,
			wantErr: true,
		},
		{
			name: "Error case - Unsupported implementation",

			cfg: Config{
				Impl: Implementation("unsupported"),
			},
			want:    nil,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewMock(t *testing.T) {
	tests := []struct {
		name string
		t    *testing.T
		want database.Adapter
	}{
		{
			name: "Nominal case",
			t:    t,
			want: databasemock.New(),
		},
		{
			name: "Error case - nil testing.T",
			t:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMock(tt.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMock() = %v, want %v", got, tt.want)
			}
		})
	}
}
