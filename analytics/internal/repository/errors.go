package repository

import "fmt"

func ErrNotExistProject(key string) error {
	return fmt.Errorf("project %s does not exist", key)
}
func ErrNotExistData(key string) error {
	return fmt.Errorf("the data of project %s for this task does not exist", key)
}

var ErrAlreadyExist = fmt.Errorf("already exist")

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

func ErrInsert(e error) error {
	return fmt.Errorf("failed to insert: %w", e)
}
