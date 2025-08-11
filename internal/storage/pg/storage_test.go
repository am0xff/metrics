package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/am0xff/metrics/internal/storage"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)

	assert.NotNil(t, pgs)
	assert.NotNil(t, pgs.db)
	assert.NotNil(t, pgs.ms)
}

func TestPGStorage_Bootstrap(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	// Expect transaction and table creation
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS gauges").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS counters").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = pgs.Bootstrap(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_SetGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	// Expect upsert query
	mock.ExpectExec("INSERT INTO gauges").
		WithArgs("test_gauge", float64(123.45)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	pgs.SetGauge(ctx, "test_gauge", storage.Gauge(123.45))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_GetGauge_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"value"}).AddRow(123.45)
	mock.ExpectQuery("SELECT value FROM gauges WHERE key = \\$1").
		WithArgs("test_gauge").
		WillReturnRows(rows)

	value, exists := pgs.GetGauge(ctx, "test_gauge")

	assert.True(t, exists)
	assert.Equal(t, storage.Gauge(123.45), value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_GetGauge_Fallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	// Set value in memory storage first
	pgs.ms.SetGauge(ctx, "test_gauge", storage.Gauge(99.99))

	// Mock DB error
	mock.ExpectQuery("SELECT value FROM gauges WHERE key = \\$1").
		WithArgs("test_gauge").
		WillReturnError(sql.ErrNoRows)

	value, exists := pgs.GetGauge(ctx, "test_gauge")

	assert.True(t, exists)
	assert.Equal(t, storage.Gauge(99.99), value)
}

func TestPGStorage_SetCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO counters").
		WithArgs("test_counter", int64(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	pgs.SetCounter(ctx, "test_counter", storage.Counter(100))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_GetCounter_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"value"}).AddRow(100)
	mock.ExpectQuery("SELECT value FROM counters WHERE key = \\$1").
		WithArgs("test_counter").
		WillReturnRows(rows)

	value, exists := pgs.GetCounter(ctx, "test_counter")

	assert.True(t, exists)
	assert.Equal(t, storage.Counter(100), value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_KeysGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"key"}).
		AddRow("gauge1").
		AddRow("gauge2")
	mock.ExpectQuery("SELECT key FROM gauges").WillReturnRows(rows)

	keys := pgs.KeysGauge(ctx)

	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "gauge1")
	assert.Contains(t, keys, "gauge2")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_KeysCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"key"}).
		AddRow("counter1").
		AddRow("counter2")
	mock.ExpectQuery("SELECT key FROM counters").WillReturnRows(rows)

	keys := pgs.KeysCounter(ctx)

	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "counter1")
	assert.Contains(t, keys, "counter2")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	pgs := NewStorage(db)
	ctx := context.Background()

	mock.ExpectPing()

	err = pgs.Ping(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
