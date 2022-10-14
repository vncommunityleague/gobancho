package bancho

import "strings"

type Member struct {
	Channel *Channel
	User    *User
	Mode    *MemberMode
}

func NewMember(client *Client, channel *Channel, username string) *Member {
	m := &Member{
		Channel: channel,
	}

	if strings.Index(username, "@") == 0 {
		m.Mode = NewMemberMode("o", "Moderator")
		username = username[1:]
	} else if strings.Index(username, "+") == 0 {
		m.Mode = NewMemberMode("v", "IRC User")
		username = username[1:]
	} else {
		m.Mode = nil
	}

	m.User = client.User(username)

	return m
}

type MemberMode struct {
	IRCLetter string
	Name      string
}

func NewMemberMode(letter, name string) *MemberMode {
	mm := &MemberMode{
		IRCLetter: letter,
		Name:      name,
	}

	return mm
}
