package rpc

import "encoding/json"

type Request struct {
	Id     string           `json:"id,omitempty"`
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params,omitempty"`
}

func (r Request) ReplyExpected() bool {
	return r.Id != ""
}

func (r Request) Reply(result *json.RawMessage) Response {
	return Response{
		RequestId: r.Id,
		Result:    result,
	}
}

func (r Request) ReplyWithError(err Error) Response {
	return Response{
		RequestId: r.Id,
		Error:     &err,
	}
}

type Response struct {
	RequestId string           `json:"requestId,omitempty"`
	Result    *json.RawMessage `json:"result,omitempty"`
	Error     *Error           `json:"error,omitempty"`
}

func (r Response) IsFailure() bool {
	return r.Error != nil
}

type ErrorCode string

const (
	ErrorCodeParseError     ErrorCode = "ParseError"
	ErrorCodeInvalidRequest ErrorCode = "InvalidRequest"
	ErrorCodeMethodNotFound ErrorCode = "MethodNotFound"
	ErrorCodeInvalidParams  ErrorCode = "InvalidParams"
	ErrorCodeInternalError  ErrorCode = "InternalError"
)

type Error struct {
	Code    ErrorCode       `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`

	cause error
}

func NewError(code ErrorCode, cause error) Error {
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
