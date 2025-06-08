package repository

import "fmt"

var ErrNotExistProject = fmt.Errorf("the project does not exist")
var ErrNotExistData = fmt.Errorf("the data of this project for this task does not exist")
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
