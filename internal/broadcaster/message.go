package broadcaster

import "time"

type Message struct {
	Id         string    `json:"id"`
	CreateTime time.Time `json:"createTime"`
	Channel    string    `json:"channel"`
	Event      string    `json:"event"`
	Payload    any       `json:"payload"`
}
