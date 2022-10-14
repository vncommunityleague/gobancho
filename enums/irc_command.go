package enums

type IRCCommand string

const (
	IRCCommandWelcome IRCCommand = "001"

	IRCCommandJoin    IRCCommand = "JOIN"
	IRCCommandQuit    IRCCommand = "QUIT"
	IRCCommandPart    IRCCommand = "PART"
	IRCCommandMode    IRCCommand = "MODE"
	IRCCommandMessage IRCCommand = "PRIVMSG"
)
