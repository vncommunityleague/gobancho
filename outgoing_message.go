package bancho

type OutgoingMessage struct {
	Client *Client

	RecipientUser    *User
	RecipientChannel *Channel
	Content          string
}

func NewOutgoingMessage(client *Client, recipientUser *User, recipientChannel *Channel, content string) *OutgoingMessage {
	msg := &OutgoingMessage{
		Client:           client,
		RecipientUser:    recipientUser,
		RecipientChannel: recipientChannel,
		Content:          content,
	}

	return msg
}

func (msg *OutgoingMessage) Send() error {
	msg.Client.MessageQueue = append(msg.Client.MessageQueue, msg)

	if len(msg.Client.MessageQueue) == 1 {
		err := msg.Client.ProcessMessageQueue()
		if err != nil {
			return err
		}
	}

	return nil
}
