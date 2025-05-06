package psql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/tezos-delegation-service/internal/model"
)

const tableDelegations = "app.delegations"

func Test_New(t *testing.T) {
	os.Setenv("GO_TESTING", "1")
	defer os.Unsetenv("GO_TESTING")

	cfg := Config{
		User:     "foo",
		DBName:   "bar",
		Host:     "localhost",
		Port:     5432,
		Password: "password",
		SSLMode:  "disable",
	}

	got, err := New(cfg)
	assert.NoError(t, err)
	p, ok := got.(*psql)
	assert.True(t, ok, "Expected *psql type")
	if ok {
		assert.NotNil(t, p.db, "DB should not be nil")
	}
}

func Test_psql_Close(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		wantErr assert.ErrorAssertionFunc
		setup   func() *sqlx.DB
	}{
		{
			name:    "Nominal case",
			wantErr: assert.NoError,
			setup: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectClose()
				return sqlx.NewDb(db, "sqlmock")
			},
		},
		{
			name:    "Error case - nil database",
			db:      nil,
			wantErr: assert.NoError,
			setup:   nil,
		},
		{
			name:    "Error case - database close error",
			wantErr: assert.Error,
			setup: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectClose().WillReturnError(errors.New("close error"))
				return sqlx.NewDb(db, "sqlmock")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *sqlx.DB
			if tt.setup != nil {
				db = tt.setup()
			} else {
				db = tt.db
			}

			p := &psql{
				db:               db,
				tableDelegations: tableDelegations,
			}

			err := p.Close()
			tt.wantErr(t, err, errors.New("Close()"))

			if db != nil {
				dbConn := db.DB
				if dbConn != nil {
					_, mock, _ := sqlmock.New()
					assert.NoError(t, mock.ExpectationsWereMet())
				}
			}
		})
	}
}

func Test_psql_GetDelegations(t *testing.T) {
	type args struct {
		ctx             context.Context
		page            uint32
		limit           uint16
		year            uint16
		maxDelegationID uint64
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
			name: "With maxDelegationID filter",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()

				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

				rows := sqlmock.NewRows([]string{"id", "delegator", "delegate", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", "delegate1", int64(1672531199), float64(1000), int64(1), createdAt)

				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM "+
					tableDelegations+
					" WHERE id <= \\$1 ORDER BY timestamp DESC LIMIT \\$2 OFFSET \\$3").
					WithArgs(int64(10), 2, 2).WillReturnRows(rows)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM " + tableDelegations).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx:             context.Background(),
				page:            2,
				limit:           2,
				year:            0,
				maxDelegationID: 10,
			},
			want: []model.Delegation{
				{
					ID:        1,
					Delegator: "delegator1",
					Delegate:  "delegate1",
					Timestamp: 1672531199,
					Amount:    1000,
					Level:     1,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			want1:   1,
			wantErr: assert.NoError,
		},
		{
			name: "With maxDelegationID filter and year",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()

				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
				endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

				rows := sqlmock.NewRows([]string{"id", "delegator", "delegate", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", "delegate1", int64(1672531199), float64(1000), int64(1), createdAt)

				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM "+
					tableDelegations+
					" WHERE timestamp >= \\$1 AND timestamp < \\$2 ORDER BY timestamp DESC LIMIT \\$3 OFFSET \\$4").
					WithArgs(startDate, endDate, 2, 2).WillReturnRows(rows)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM "+tableDelegations+
					" WHERE timestamp >= \\$1 AND timestamp < \\$2").
					WithArgs(startDate, endDate).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx:             context.Background(),
				page:            2,
				limit:           2,
				year:            2023,
				maxDelegationID: 10,
			},
			want: []model.Delegation{
				{
					ID:        1,
					Delegator: "delegator1",
					Delegate:  "delegate1",
					Timestamp: 1672531199,
					Amount:    1000,
					Level:     1,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want1:   1,
			wantErr: assert.NoError,
		},
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()

				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

				rows := sqlmock.NewRows([]string{"id", "delegator", "delegate", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", "delegate1", int64(1672531199), float64(1000), int64(1), createdAt).
					AddRow(2, "delegator2", "", int64(1672531200), float64(2000), int64(2), createdAt)

				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM "+
					tableDelegations+" ORDER BY timestamp DESC LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).WillReturnRows(rows)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM " + tableDelegations).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx:             context.Background(),
				page:            1,
				limit:           2,
				year:            0,
				maxDelegationID: 0,
			},
			want: []model.Delegation{
				{
					ID:        1,
					Delegator: "delegator1",
					Delegate:  "delegate1",
					Timestamp: 1672531199,
					Amount:    1000, Level: 1,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        2,
					Delegator: "delegator2",
					Delegate:  "",
					Timestamp: 1672531200,
					Amount:    2000,
					Level:     2,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want1:   2,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM "+
					tableDelegations+
					" ORDER BY timestamp DESC LIMIT \\$1 OFFSET \\$2").
					WithArgs(2, 0).WillReturnError(context.Canceled)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				page:            1,
				limit:           2,
				year:            0,
				maxDelegationID: 0,
			},
			want:    nil,
			want1:   0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			got, err := p.GetDelegations(tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year, tt.args.maxDelegationID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetDelegations(%v, %v, %v, %v, %v)",
				tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year, tt.args.maxDelegationID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetDelegations(%v, %v, %v, %v, %v)",
				tt.args.ctx, tt.args.page, tt.args.limit, tt.args.year, tt.args.maxDelegationID)
		})
	}
}

func Test_psql_GetHighestBlockLevel(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		want    uint64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM " + tableDelegations).
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
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM " + tableDelegations).
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
				mock.ExpectQuery("SELECT COALESCE\\(MAX\\(level\\), 0\\) FROM " + tableDelegations).
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
				db:               tt.db,
				tableDelegations: tableDelegations,
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
				createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{"id", "delegator", "delegate", "timestamp", "amount", "level", "created_at"}).
					AddRow(1, "delegator1", "delegate1", int64(1672531199), float64(1000), int64(1), createdAt)
				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM " +
					tableDelegations + " ORDER BY level DESC LIMIT 1").
					WillReturnRows(rows)
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx: context.Background(),
			want: &model.Delegation{
				ID:        1,
				Delegator: "delegator1",
				Delegate:  "delegate1",
				Timestamp: 1672531199,
				Amount:    1000,
				Level:     1,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM " +
					tableDelegations + " ORDER BY level DESC LIMIT 1").
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
				mock.ExpectQuery("SELECT id, delegator, delegate, timestamp, amount, level, created_at FROM " +
					tableDelegations + " ORDER BY level DESC LIMIT 1").
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
				db:               tt.db,
				tableDelegations: tableDelegations,
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
				mock.ExpectPing()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			err := p.Ping()
			tt.wantErr(t, err, fmt.Sprintf("Ping()"))
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
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "delegate1", int64(1672531199), float64(1000), int64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegation: &model.Delegation{
					Delegator: "delegator1",
					Delegate:  "delegate1",
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
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "delegate1", 1672531199, 1000, 1).
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
					Delegate:  "delegate1",
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
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "delegate1", 1672531199, 1000, 1).
					WillReturnError(fmt.Errorf("database error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegation: &model.Delegation{
					Delegator: "delegator1",
					Delegate:  "delegate1",
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
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			tt.wantErr(t, p.SaveDelegation(tt.args.ctx, tt.args.delegation),
				fmt.Sprintf("SaveDelegation(%v, %v)", tt.args.ctx, tt.args.delegation))
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
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "delegate2", int64(1672531199), float64(1000), int64(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator2", "delegate3", int64(1672531200), float64(2000), int64(2)).
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegations: []*model.Delegation{
					{Delegator: "delegator1", Delegate: "delegate2", Timestamp: 1672531199, Amount: 1000, Level: 1},
					{Delegator: "delegator2", Delegate: "delegate3", Timestamp: 1672531200, Amount: 2000, Level: 2},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - context canceled",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "", 1672531199, 1000, 1).
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
					{Delegator: "delegator1", Delegate: "", Timestamp: 1672531199, Amount: 1000, Level: 1},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Error case - database error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO "+tableDelegations).
					WithArgs("delegator1", "", 1672531199, 1000, 1).
					WillReturnError(fmt.Errorf("database error"))
				mock.ExpectRollback()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				delegations: []*model.Delegation{
					{Delegator: "delegator1", Delegate: "", Timestamp: 1672531199, Amount: 1000, Level: 1},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			tt.wantErr(t, p.SaveDelegations(tt.args.ctx, tt.args.delegations),
				fmt.Sprintf("SaveDelegations(%v, %v)", tt.args.ctx, tt.args.delegations))
		})
	}
}

func Test_psql_GetLastSyncedRewardCycle(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(last_synced_level, 0\\) AS cycle FROM app.sync_state WHERE source = 'rewards' LIMIT 1").
					WillReturnRows(sqlmock.NewRows([]string{"cycle"}).AddRow(10))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    10,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT COALESCE\\(last_synced_level, 0\\) AS cycle FROM app.sync_state WHERE source = 'rewards' LIMIT 1").
					WillReturnError(fmt.Errorf("query error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			got, err := p.GetLastSyncedRewardCycle(tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetLastSyncedRewardCycle(%v)", tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetLastSyncedRewardCycle(%v)", tt.ctx)
		})
	}
}

func Test_psql_GetActiveDelegators(t *testing.T) {
	const tableDelegations = "app.delegations"
	
	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		want    []model.WalletAddress
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT DISTINCT delegator AS address FROM "+tableDelegations+" WHERE amount > 0 ORDER BY delegator").
					WillReturnRows(sqlmock.NewRows([]string{"address"}).AddRow("tz1delegator1").AddRow("tz1delegator2"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    []model.WalletAddress{"tz1delegator1", "tz1delegator2"},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT DISTINCT delegator AS address FROM "+tableDelegations+" WHERE amount > 0 ORDER BY delegator").
					WillReturnError(fmt.Errorf("query error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			got, err := p.GetActiveDelegators(tt.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetActiveDelegators(%v)", tt.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetActiveDelegators(%v)", tt.ctx)
		})
	}
}

func Test_psql_GetBakerForDelegatorAtCycle(t *testing.T) {
	const tableDelegations = "app.delegations"
	
	tests := []struct {
		name      string
		db        *sqlx.DB
		ctx       context.Context
		delegator model.WalletAddress
		cycle     int
		want      model.WalletAddress
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				// Nous ne testons pas exactement l'approximation du timestamp, mais la requête elle-même
				mock.ExpectQuery("SELECT delegate FROM "+tableDelegations+" WHERE delegator = \\$1 AND timestamp <= \\$2 ORDER BY timestamp DESC LIMIT 1").
					WithArgs("tz1delegator1", mock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"delegate"}).AddRow("tz1baker1"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:       context.Background(),
			delegator: "tz1delegator1",
			cycle:     10,
			want:      "tz1baker1",
			wantErr:   assert.NoError,
		},
		{
			name: "Error case - query error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT delegate FROM "+tableDelegations+" WHERE delegator = \\$1 AND timestamp <= \\$2 ORDER BY timestamp DESC LIMIT 1").
					WithArgs("tz1delegator1", mock.AnyArg()).
					WillReturnError(fmt.Errorf("query error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:       context.Background(),
			delegator: "tz1delegator1",
			cycle:     10,
			want:      "",
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:               tt.db,
				tableDelegations: tableDelegations,
			}
			got, err := p.GetBakerForDelegatorAtCycle(tt.ctx, tt.delegator, tt.cycle)
			if !tt.wantErr(t, err, fmt.Sprintf("GetBakerForDelegatorAtCycle(%v, %v, %v)", tt.ctx, tt.delegator, tt.cycle)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetBakerForDelegatorAtCycle(%v, %v, %v)", tt.ctx, tt.delegator, tt.cycle)
		})
	}
}

func Test_psql_SaveRewards(t *testing.T) {
	const tableRewards = "app.rewards"
	
	type args struct {
		ctx     context.Context
		rewards []model.Reward
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
				mock.ExpectExec("INSERT INTO "+tableRewards).
					WithArgs("tz1delegator1", "tz1baker1", 10, 5.5, int64(1672531199)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO "+tableRewards).
					WithArgs("tz1delegator2", "tz1baker2", 10, 3.3, int64(1672531200)).
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				rewards: []model.Reward{
					{RecipientAddress: "tz1delegator1", SourceAddress: "tz1baker1", Cycle: 10, Amount: 5.5, Timestamp: 1672531199},
					{RecipientAddress: "tz1delegator2", SourceAddress: "tz1baker2", Cycle: 10, Amount: 3.3, Timestamp: 1672531200},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Error case - transaction begin error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				rewards: []model.Reward{
					{RecipientAddress: "tz1delegator1", SourceAddress: "tz1baker1", Cycle: 10, Amount: 5.5, Timestamp: 1672531199},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "Error case - insert error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO "+tableRewards).
					WithArgs("tz1delegator1", "tz1baker1", 10, 5.5, int64(1672531199)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
				return sqlx.NewDb(db, "sqlmock")
			}(),
			args: args{
				ctx: context.Background(),
				rewards: []model.Reward{
					{RecipientAddress: "tz1delegator1", SourceAddress: "tz1baker1", Cycle: 10, Amount: 5.5, Timestamp: 1672531199},
				},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db:          tt.db,
				tableRewards: tableRewards,
			}
			tt.wantErr(t, p.SaveRewards(tt.args.ctx, tt.args.rewards),
				fmt.Sprintf("SaveRewards(%v, %v)", tt.args.ctx, tt.args.rewards))
		})
	}
}

func Test_psql_SaveLastSyncedRewardCycle(t *testing.T) {
	tests := []struct {
		name    string
		db      *sqlx.DB
		ctx     context.Context
		cycle   int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Nominal case",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("INSERT INTO app.sync_state").
					WithArgs("rewards", 10, mock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			cycle:   10,
			wantErr: assert.NoError,
		},
		{
			name: "Error case - insert error",
			db: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("INSERT INTO app.sync_state").
					WithArgs("rewards", 10, mock.AnyArg()).
					WillReturnError(fmt.Errorf("insert error"))
				return sqlx.NewDb(db, "sqlmock")
			}(),
			ctx:     context.Background(),
			cycle:   10,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &psql{
				db: tt.db,
			}
			tt.wantErr(t, p.SaveLastSyncedRewardCycle(tt.ctx, tt.cycle),
				fmt.Sprintf("SaveLastSyncedRewardCycle(%v, %v)", tt.ctx, tt.cycle))
		})
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("RESULT") == "ok" {
		os.Exit(0)
	}
	os.Exit(1)
}
