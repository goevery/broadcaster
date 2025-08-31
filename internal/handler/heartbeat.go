package handler

import "time"

type HeartbeatResponse struct {
	Timestamp int64 `json:"timestamp"`
}

type HeartbeatHandler struct{}

func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

func (h *HeartbeatHandler) Handle() HeartbeatResponse {
	return HeartbeatResponse{
		Timestamp: time.Now().Unix(),
	}
}
