package bancho

import "github.com/vncommunityleague/gobancho/enums"

type Channel struct {
	Client *Client

	Name     string
	Topic    string
	Type     enums.ChannelType
	IsJoined bool
	Members  map[string]*Member

	Lobby *Lobby
}

func NewChannel(client *Client, channelName string) *Channel {
	channel := &Channel{
		Client:   client,
		Name:     channelName,
		Topic:    "",
		IsJoined: false,
		Members:  make(map[string]*Member),
	}

	return channel
}

func (ch *Channel) Join() {
	ch.Client.Send("JOIN " + ch.Name)
}
