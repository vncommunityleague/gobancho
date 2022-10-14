package bancho

import "time"

type Message struct {
	Author    *User
	Channel   *Channel
	Content   string
	Timestamp *time.Time
	IsAction  bool
}

func NewMessage(author *User, channel *Channel, content string, timestamp *time.Time, isAction bool) *Message {
	msg := &Message{
		Author:    author,
		Channel:   channel,
		Content:   content,
		Timestamp: timestamp,
		IsAction:  isAction,
	}

	return msg
}
