package utils

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"os"
	"syscall"
	"time"
)

func Do(ctx context.Context, f func() error) error {
	intervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var err error
	for _, interval := range intervals {
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return ctx.Err()
		}

		err = f()
		if !isRetriable(err) {
			break
		}
	}

	return err
}

func isRetriable(err error) bool {
	if pgErr := asPgError(err); pgErr != nil {
		return isPgErrorRetryable(pgErr)
	}
	if isFileError(err) {
		return true
	}

	return false
}

func isFileError(err error) bool {
	var pe *os.PathError
	if errors.As(err, &pe) {
		switch pe.Err {
		case syscall.EAGAIN,
			syscall.EBUSY:
			return true
		}
	}
	return false
}

func asPgError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr
	}
	return nil
}

func isPgErrorRetryable(pgErr *pgconn.PgError) bool {
	code := pgErr.Code
	if pgerrcode.IsConnectionException(code) {
		return true
	}

	return false
}
