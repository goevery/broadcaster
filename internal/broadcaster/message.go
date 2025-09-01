package broadcaster

import "time"

type Message struct {
	Id         string    `json:"id"`
	CreateTime time.Time `json:"createTime"`
	ChannelId  string    `json:"channelId"`
	Payload    any       `json:"payload"`
}
