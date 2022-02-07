package model

const (
	// ErrCommandNotConverted error command not recognize.
	ErrCommandNotConverted = Error("command not converted")
	// ErrUserNotFound error user not found.
	ErrUserNotFound = Error("user not found")
	// ErrFoundTwoUsers error found two user account for one user.
	ErrFoundTwoUsers = Error("found two users")

	ErrNotSelectedLanguage = Error("not selected language")

	// ErrScanSqlRow error scan sql row.
	ErrScanSqlRow = Error("failed scan sql row")
)

type Error string

func (e Error) Error() string {
	return string(e)
}
