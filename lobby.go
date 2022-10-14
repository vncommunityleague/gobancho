package bancho

type Lobby struct {
	Client    *Client
	Channel   *Channel
	ID        uint
	Name      string
	BeatmapID uint
}
