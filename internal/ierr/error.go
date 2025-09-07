package ierr

import "encoding/json"

type ErrorCode string

const (
	ErrorCodeInvalidArgument    ErrorCode = "InvalidArgument"
	ErrorCodeNotFound           ErrorCode = "NotFound"
	ErrorCodeAlreadyExists      ErrorCode = "AlreadyExists"
	ErrorCodeFailedPrecondition ErrorCode = "FailedPrecondition"
	ErrorCodePermissionDenied   ErrorCode = "PermissionDenied"
	ErrorCodeUnauthenticated    ErrorCode = "Unauthenticated"
	ErrorCodeInternal           ErrorCode = "Internal"
)

type Error struct {
	Code    ErrorCode       `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`

	cause error
}

func New(code ErrorCode, cause error) Error {
	return Error{
		Code:    code,
		Message: cause.Error(),
		cause:   cause,
	}
}

func (e Error) Error() string {
	return string(e.Code) + ": " + e.cause.Error()
}

func (e Error) Unwrap() error {
	return e.cause
}
