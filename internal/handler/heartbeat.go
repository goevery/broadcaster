package handler

import "time"

type HeartbeatResponse struct {
	Timestamp time.Time `json:"timestamp"`
}

type HeartbeatHandler struct{}

func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

func (h *HeartbeatHandler) Handle() HeartbeatResponse {
	return HeartbeatResponse{
		Timestamp: time.Now(),
	}
}
