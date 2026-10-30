// Package httpx contains HTTP error values mapped by the central Fiber handler.
package httpx

type Error struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func NewError(status int, code, message string, cause error) *Error {
	return &Error{Status: status, Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string { return e.Code }

func (e *Error) Unwrap() error { return e.Cause }
