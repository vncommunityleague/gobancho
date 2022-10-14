package bancho

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vncommunityleague/gobancho/enums"
	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"
)

var IgnoredSplits = []string{
	"312", // Whois server info (useless on Bancho)
	"333", // Time when topic was set
	"366", // End of NAMES reply
	"372", // MOTD
	"375", // MOTD Begin
	"376", // MOTD End
}

type Config struct {
	Username   string
	Password   string
	Host       string
	Port       uint16
	APIKey     string
	BotAccount bool
}

var DefaultConfig = Config{
	Username:   "",
	Password:   "",
	Host:       "irc.ppy.sh",
	Port:       6667,
	APIKey:     "",
	BotAccount: false,
}

type Client struct {
	Username   string
	Password   string
	Host       string
	Port       uint16
	APIKey     string
	BotAccount bool

	ConnectionState enums.ConnectionState
	Users           map[string]*User
	Channels        map[string]*Channel
	MessageQueue    []*OutgoingMessage
	RateLimiter     *rate.Limiter

	OnMessageReceived OnMessageReceivedHandler
}

type OnMessageReceivedHandler func(*Client)

func New(config ...Config) *Client {
	cfg := DefaultConfig

	if len(config) > 0 {
		cfg = config[0]

		if cfg.Host == "" {
			cfg.Host = DefaultConfig.Host
		}

		if cfg.Port == 0 {
			cfg.Port = DefaultConfig.Port
		}
	}

	client := &Client{
		Username:   cfg.Username,
		Password:   cfg.Password,
		Host:       cfg.Host,
		Port:       cfg.Port,
		APIKey:     cfg.APIKey,
		BotAccount: cfg.BotAccount,

		Users:        make(map[string]*User),
		Channels:     make(map[string]*Channel),
		MessageQueue: []*OutgoingMessage{},
		RateLimiter:  rate.NewLimiter(rate.Every(2*time.Second), 1),
	}

	return client
}

var (
	conn net.Conn
)

func (c *Client) Connect() {
	if c.ConnectionState == enums.ConnectionStateConnected {
		return
	}

	var err error
	conn, err = net.DialTimeout("tcp", c.Host+":"+strconv.Itoa(int(c.Port)), time.Minute)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	c.ConnectionState = enums.ConnectionStateConnecting

	c.Send("PASS " + c.Password)
	c.Send("USER " + c.Username + " 0 * :" + c.Username)
	c.Send("NICK " + c.Username)
}

func (c *Client) Disconnect() {
	c.Send("QUIT")
	err := conn.Close()
	if err != nil {
		log.Error().Err(err).Send()
	}

	c.ConnectionState = enums.ConnectionStateDisconnected
}

func (c *Client) Send(message string) {
	if (c.ConnectionState == enums.ConnectionStateConnected) || (c.ConnectionState == enums.ConnectionStateConnecting) {
		if _, err := fmt.Fprintf(conn, "%s\r\n", message); err != nil {
			log.Error().Err(err).Send()
			return
		}

		log.Debug().Msg(message)
	}
}

func (c *Client) Listen() {
	tp := textproto.NewReader(bufio.NewReader(conn))

	for {
		data, err := tp.ReadLine()
		if err != nil {
			panic(err)
		}

		c.HandleIRCCommand(data)
	}
}

func (c *Client) HandleIRCCommand(message string) {
	args := strings.Split(message, " ")

	if slices.Contains(IgnoredSplits, args[1]) {
		return
	}

	if args[1] == "PART" || args[1] == "MODE" || args[1] == "353" {
		return
	}

	if args[0] == "PING" {
		c.Send("PONG " + strings.Join(args[1:], " "))
		return
	}

	switch args[1] {
	case string(enums.IRCCommandWelcome):
		{
			c.ConnectionState = enums.ConnectionStateConnected
		}

	case string(enums.IRCCommandMessage):
		{
			user := c.User((args[0])[1:strings.Index(args[0], "!")])
			content := strings.Join(args[3:], " ")[1:]
			channel, _ := c.Channel(args[2])
			now := time.Now()

			msg := NewMessage(user, channel, content, &now, false)

			log.Info().Str("user", msg.Author.IRCUsername).Str("content", msg.Content).Send()
		}

	case string(enums.IRCCommandJoin):
		{
			user := c.User((args[0])[1:strings.Index(args[0], "!")])
			channel, _ := c.Channel((args[2])[1:])

			member := NewMember(c, channel, user.IRCUsername)
			channel.Members[user.IRCUsername] = member

			if user.IsClient() {
				channel.IsJoined = true
			}
		}

	case string(enums.IRCCommandQuit):
		{
			username := (args[0])[1:strings.Index(args[0], "!")]
			user := c.Users[username]

			if user != nil {
				for _, channel := range c.Channels {
					delete(channel.Members, user.IRCUsername)
				}
			}
		}
	}
}

func (c *Client) Self() *User {
	return c.User(c.Username)
}

func (c *Client) User(username string) *User {
	username = strings.Split(strings.ReplaceAll(username, " ", "_"), "\n")[0]
	user := c.Users[strings.ToLower(username)]

	if user == nil {
		user = NewUser(c, username)
		c.Users[strings.ToLower(username)] = user
	}

	return user
}

func (c *Client) Channel(channelName string) (*Channel, error) {
	if strings.Index(channelName, "#") != 0 || strings.Index(channelName, ",") != -1 {
		return nil, errors.New("invalid channel name")
	}

	channel := c.Channels[channelName]

	if channel == nil {
		channel = NewChannel(c, channelName)
		c.Channels[channelName] = channel
	}

	return channel, nil
}

func (c *Client) CreateLobby(name string) (*Channel, error) {
	if c.APIKey == "" {
		return nil, errors.New("missing API key")
	}

	if name == "" || strings.TrimSpace(name) == "" {
		return nil, errors.New("missing lobby name")
	}

	name = strings.TrimSpace(name)

	banchoBot := c.User("Hoaq")
	_, err := banchoBot.Send("!mp make " + name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ProcessMessageQueue Process client messages
func (c *Client) ProcessMessageQueue() error {
	for len(c.MessageQueue) > 0 {
		if c.ConnectionState != enums.ConnectionStateConnected {
			return errors.New("client is currently disconnected")
		}

		err := c.RateLimiter.Wait(context.Background())
		if err != nil {
			return err
		}

		msg := c.MessageQueue[0]
		name := ""
		content := strings.Split(msg.Content, "\n")[0]
		now := time.Now()

		// TODO:
		if msg.RecipientUser != nil && msg.RecipientChannel == nil {
			name = strings.Split(strings.ReplaceAll(msg.RecipientUser.IRCUsername, " ", "_"), "\n")[0]
			//NewMessage(c.Self(), msg.RecipientUser, content)
		} else if msg.RecipientChannel != nil && msg.RecipientUser == nil {
			name = strings.Split(strings.ReplaceAll(msg.RecipientChannel.Name, " ", "_"), "\n")[0]
			NewMessage(c.Self(), msg.RecipientChannel, content, &now, false)
		}

		c.Send("PRIVMSG " + name + " :" + content)

		// Shift message queue
		c.MessageQueue = c.MessageQueue[1:]
	}

	return nil
}
