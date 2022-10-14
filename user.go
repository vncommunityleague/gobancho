package bancho

import (
	"strings"
	"time"
)

type User struct {
	Client      *Client
	IRCUsername string

	ID                 uint
	Username           string
	Country            string
	JoinDate           *time.Time
	PerformancePoint   float64
	GlobalRank         uint
	CountryRank        uint
	TotalScore         uint
	RankedScore        uint
	Accuracy           float64
	Count300           uint
	Count100           uint
	Count50            uint
	PlayCount          uint
	TotalSecondsPlayed uint
	Level              float64
	CountRankSS        uint
	CountRankS         uint
	CountRankA         uint
}

func NewUser(client *Client, username string) *User {
	u := &User{
		Client:      client,
		IRCUsername: username,
	}

	return u
}

func (u *User) IsClient() bool {
	return strings.ToLower(u.Client.Username) == strings.ToLower(u.IRCUsername)
}

func (u *User) Send(content string) (*Message, error) {
	err := (NewOutgoingMessage(u.Client, u, nil, content)).Send()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
