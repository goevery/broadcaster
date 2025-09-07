package handler

import (
	"encoding/json"

	"github.com/goevery/broadcaster/internal/ierr"
)

type Request struct {
	Id     int              `json:"id,omitempty"`
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params,omitempty"`
}

func NewNotification(method string, params *json.RawMessage) Request {
	return Request{
		Method: method,
		Params: params,
	}
}

func (r Request) ReplyExpected() bool {
	return r.Id != 0
}

func (r Request) Reply(result *json.RawMessage) Response {
	return Response{
		RequestId: r.Id,
		Result:    result,
	}
}

func (r Request) ReplyWithError(err ierr.Error) Response {
	return Response{
		RequestId: r.Id,
		Error:     &err,
	}
}

type Response struct {
	RequestId int              `json:"requestId,omitempty"`
	Result    *json.RawMessage `json:"result,omitempty"`
	Error     *ierr.Error      `json:"error,omitempty"`
}

func (r Response) IsFailure() bool {
	return r.Error != nil
}
