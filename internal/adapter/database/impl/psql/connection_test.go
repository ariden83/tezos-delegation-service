package psql

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_initConnection(t *testing.T) {
	// db, mock, _ := sqlmock.New()
	// mock.ExpectPing().WillReturnError(nil)
	// dbInstance := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name    string
		config  Config
		want    *sqlx.DB
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			config: Config{
				Driver:   "sqlmock",
				User:     "foo",
				DBName:   "bar",
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				SSLMode:  "disable",
			},
			want:    &sqlx.DB{},
			wantErr: assert.NoError,
		},
		{
			name: "Error creating directory",
			config: Config{
				Driver:   "sqlmock",
				User:     "foo",
				DBName:   "bar",
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				SSLMode:  "disable",
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Error running migrations",
			config: Config{
				Driver: "postgres",
				Host:   "invalid_dsn",
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Error connecting to database",
			config: Config{
				Driver:   "postgres",
				User:     "invalid_user",
				DBName:   "invalid_db",
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				SSLMode:  "disable",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initConnection(tt.config)
			if !tt.wantErr(t, err, fmt.Sprintf("initConnection(%v)", tt.config)) {
				return
			}
			assert.Equalf(t, tt.want, got, "initConnection(%v)", tt.config)
		})
	}
}

func Test_runSqitchMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Échec du lancement du conteneur PostgreSQL: %v", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("Échec de l'arrêt du conteneur: %v", err)
		}
	}()

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Impossible d'obtenir le port mappé: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Impossible d'obtenir l'hôte: %v", err)
	}

	tests := []struct {
		name    string
		config  Config
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case - real PostgreSQL",
			config: Config{
				Driver:   "postgres",
				Host:     host,
				Port:     port.Int(),
				User:     "testuser",
				Password: "testpass",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - missing script",
			config: Config{
				Driver:   "postgres",
				Host:     host,
				Port:     port.Int(),
				User:     "testuser",
				Password: "testpass",
				DBName:   "nonexistent",
				SSLMode:  "disable",
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, runSqitchMigrations(tt.config), fmt.Sprintf("runSqitchMigrations(%v)", tt.config))
		})
	}
}
