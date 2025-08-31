package handler

type ErrorCode string

const (
	ErrorCodeInvalidArgument ErrorCode = "invalid_argument"
)

type Error struct {
	code  ErrorCode
	cause error
}

func NewError(code ErrorCode, cause error) Error {
	return Error{
		code:  code,
		cause: cause,
	}
}

func (e Error) Code() ErrorCode {
	return e.code
}

func (e Error) Error() string {
	return string(e.code) + ": " + e.cause.Error()
}

func (e Error) Unwrap() error {
	return e.cause
}
