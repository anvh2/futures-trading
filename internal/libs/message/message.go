package message

import "time"

type Message struct {
	expire time.Time
	Offset int64
	Data   interface{}
}
