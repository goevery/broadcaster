package server

import "net/http"

type OriginChecker struct {
}

func NewOriginChecker() *OriginChecker {
	return &OriginChecker{}
}

func (o *OriginChecker) Check(r *http.Request) bool {
	return true
}
