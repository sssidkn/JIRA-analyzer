package repository

import "fmt"

var ErrNotExist = fmt.Errorf("not exist")

func ErrExistence(err error) error {
	return fmt.Errorf("failed to check existense: %w", err)
}

func ErrScan(err error) error {
	return fmt.Errorf("failed to scan: %w", err)
}

func ErrSelect(err error) error {
	return fmt.Errorf("failed to select: %w", err)
}

func ErrDelete(err error) error {
	return fmt.Errorf("failed to delete: %w", err)
}

func ErrBeginTransaction(e error) error {
	return fmt.Errorf("failed to begin transaction: %w", e)
}

func ErrCommitTransaction(e error) error {
	return fmt.Errorf("failed to commit transaction: %w", e)
}
