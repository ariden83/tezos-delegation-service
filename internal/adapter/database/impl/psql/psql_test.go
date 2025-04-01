package psql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/adapter/database"
	"github.com/tezos-delegation-service/internal/model"
)

func Test_New(t *testing.T) {

	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name    string
		cfg     Config
		want    database.Adapter
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			cfg: Config{
				Driver:   "sqlmock",
				User:     "foo",
				DBName:   "bar",
				Host:     "localhost",
				Port:     5432,
				Password: "password",
				SSLMode:  "disable",
			},
			want: &psql{
				db: sqlxDB,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - invalid driver",
			cfg: Config{
				Driver:   "invalid_driver",
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
			name: "Error case - empty host",
			cfg: Config{
				Driver: "sqlmock",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.cfg)
			if !tt.wantErr(t, err, fmt.Sprintf("New(%v)", tt.cfg)) {
				return
			}
			assert.Equalf(t, tt.want, got, "New(%v)", tt.cfg)
		})
	}
}

func Test_psql_Close(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name    string
		db      *sqlx.DB
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Nominal case",
			db:      sqlxDB,
			wantErr: assert.NoError,
		},
		{
			name:    "Error case - nil database",
			db:      nil,
			wantErr: assert.Error,
		},
		{
			name: "Error case - database close error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectClose().WillReturnError(fmt.Errorf("close error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			tt.wantErr(t, p.Close(), fmt.Sprintf("Close()"))
		})
	}
}

func Test_psql_CountDelegations(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	type args struct {
		ctx  context.Context
		year int
	}
	tests := []struct {
		name    string
		db      *sqlx.DB
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db:   sqlxDB,
			args: args{
				ctx:  context.Background(),
				year: 2023,
			},
			want:    10,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - invalid year",
			db:   sqlxDB,
			args: args{
				ctx:  context.Background(),
				year: -1,
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "Error case - context canceled",
			db:   sqlxDB,
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				year: 2023,
			},
			want:    0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			got, err := p.CountDelegations(tt.args.ctx, tt.args.year)
			if !tt.wantErr(t, err, fmt.Sprintf("CountDelegations(%v, %v)", tt.args.ctx, tt.args.year)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CountDelegations(%v, %v)", tt.args.ctx, tt.args.year)
		})
	}
}

func Test_psql_GetDelegations(t *testing.T) {
	type args struct {
		ctx   context.Context
		page  int
		limit int
		year  int
	}
	tests := []struct {
		name    string
		db      *sqlx.DB
		args    args
		want    []model.Delegation
		want1   int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				rows := sqlmock.NewRows([]string{"id", "delegator", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", 1672531199, 1000, 1, time.Now()).
					AddRow(2, "delegator2", 1672531200, 2000, 2, time.Now())
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY timestamp DESC LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).WillReturnRows(rows)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM delegations").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx:   context.Background(),
				page:  1,
				limit: 2,
				year:  0,
			},
			want: []model.Delegation{
				{ID: 1, Delegator: "delegator1", Timestamp: 1672531199, Amount: 1000, Level: 1},
				{ID: 2, Delegator: "delegator2", Timestamp: 1672531200, Amount: 2000, Level: 2},
			},
			want1:   2,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - invalid year",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations WHERE timestamp >= \\$1 AND timestamp < \\$2 ORDER BY timestamp DESC LIMIT \\$3 OFFSET \\$4").
					WithArgs(int64(-62135596800), int64(-62135596800), 2, 0).WillReturnError(fmt.Errorf("invalid year"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx:   context.Background(),
				page:  1,
				limit: 2,
				year:  -1,
			},
			want:    nil,
			want1:   0,
			wantErr: assert.Error,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, timestamp, amount, level, created_at FROM delegations ORDER BY timestamp DESC LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).WillReturnError(context.Canceled)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				page:  1,
				limit: 2,
				year:  0,
			},
			want:    nil,
			want1:   0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			got, got1, err := p.GetDelegations(tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year)
			if !tt.wantErr(t, err, fmt.Sprintf("GetDelegations(%v, %v, %v, %v)", tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetDelegations(%v, %v, %v, %v)", tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year)
			assert.Equalf(t, tt.want1, got1, "GetDelegations(%v, %v, %v, %v)", tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year)
		})
	}
}

func Test_psql_GetHighestBlockLevel(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		want    int64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM delegations").
					WillReturnRows(sqlmock.NewRows([]string{"level"}).AddRow(100))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    100,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM delegations").
					WillReturnError(fmt.Errorf("query error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM delegations").
					WillReturnError(context.Canceled)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			want:    0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			got, err := p.GetHighestBlockLevel(tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetHighestBlockLevel(%v)", tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetHighestBlockLevel(%v)", tt.ctx)
		})
	}
}

func Test_psql_GetLatestDelegation(t *testing.T) {
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
				rows := sqlmock.NewRows([]string{"id", "delegator", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", 1672531199, 1000, 1, time.Now())
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
			},
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
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			got, err := p.GetLatestDelegation(tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetLatestDelegation(%v)", tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetLatestDelegation(%v)", tt.ctx)
		})
	}
}

func Test_psql_Ping(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectPing().WillReturnError(nil)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			wantErr: assert.NoError,
		},
		{
			name: "Error case - ping error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectPing().WillReturnError(fmt.Errorf("ping error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			wantErr: assert.Error,
		},
		{
			name:    "Error case - nil database",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			tt.wantErr(t, p.Ping(), fmt.Sprintf("Ping()"))
		})
	}
}

func Test_psql_SaveDelegation(t *testing.T) {
	type args struct {
		ctx        context.Context
		delegation *model.Delegation
	}
	tests := []struct {
		name    string
		db      *sqlx.DB
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegation: &model.Delegation{
					Delegator: "delegator1",
					Timestamp: 1672531199,
					Amount:    1000,
					Level:     1,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnError(context.Canceled)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				delegation: &model.Delegation{
					Delegator: "delegator1",
					Timestamp: 1672531199,
					Amount:    1000,
					Level:     1,
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Error case - database error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnError(fmt.Errorf("database error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegation: &model.Delegation{
					Delegator: "delegator1",
					Timestamp: 1672531199,
					Amount:    1000,
					Level:     1,
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			tt.wantErr(t, p.SaveDelegation(tt.args.ctx, tt.args.delegation), fmt.Sprintf("SaveDelegation(%v, %v)", tt.args.ctx, tt.args.delegation))
		})
	}
}

func Test_psql_SaveDelegations(t *testing.T) {
	type args struct {
		ctx         context.Context
		delegations []*model.Delegation
	}
	tests := []struct {
		name    string
		db      *sqlx.DB
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator2", 1672531200, 2000, 2).
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegations: []*model.Delegation{
					{Delegator: "delegator1", Timestamp: 1672531199, Amount: 1000, Level: 1},
					{Delegator: "delegator2", Timestamp: 1672531200, Amount: 2000, Level: 2},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnError(context.Canceled)
				mock.ExpectRollback()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				delegations: []*model.Delegation{
					{Delegator: "delegator1", Timestamp: 1672531199, Amount: 1000, Level: 1},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Error case - database error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO delegations").
					WithArgs("delegator1", 1672531199, 1000, 1).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegations: []*model.Delegation{
					{Delegator: "delegator1", Timestamp: 1672531199, Amount: 1000, Level: 1},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			tt.wantErr(t, p.SaveDelegations(tt.args.ctx, tt.args.delegations), fmt.Sprintf("SaveDelegations(%v, %v)", tt.args.ctx, tt.args.delegations))
		})
	}
}
