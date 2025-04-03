package psql

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* func Test_initConnection(t *testing.T) {

	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess"}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "RESULT=ok"}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}

	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		want    *model.Delegation
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{"id", "delegator", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", int64(1672531199), float64(1000), int64(1), createdAt)
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY level DESC LIMIT 1").
					WillReturnRows(rows)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx: context.Background(),
			want: &model.Delegation{
				ID:        1,
				Delegator: "delegator1",
				Timestamp: 1672531199,
				Amount:    1000,
				Level:     1,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: assert.NoError,
		},
		{
			name: "No rows found",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				// Return empty result set (no rows)
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY level DESC LIMIT 1").
					WillReturnRows(sqlmock.NewRows([]string{"id", "delegator", "timestamp", "amount", "level", "created_at"}))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY level DESC LIMIT 1").
					WillReturnError(fmt.Errorf("query error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY level DESC LIMIT 1").
					WillReturnError(context.Canceled)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Driver:   "sqlmock",
				User:     "invalid_user",
				DBName:   "invalid_db",
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				SSLMode:  "disable"}
			got, err := initConnection(cfg)
			if !tt.wantErr(t, err, fmt.Sprintf("initConnection(%v)", cfg)) {
				return
			}
			assert.Equalf(t, tt.want, got, "initConnection(%v)", cfg)
		})
	}
} */

func Test_runSqitchMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess"}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "RESULT=ok"}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}

	config := Config{
		DBMigrateFile: "../../../../../scripts/db_migrate.sh",
		Driver:        "sqlmock",
		Host:          "localhost",
		Port:          5432,
		User:          "testuser",
		Password:      "testpass",
		DBName:        "testdb",
		SSLMode:       "disable",
	}

	assert.NoError(t, runSqitchMigrations(config), fmt.Sprintf("runSqitchMigrations(%v)", config))
}
