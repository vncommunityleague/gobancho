package enums

type ConnectionState int

const (
	ConnectionStateDisconnected ConnectionState = iota
	ConnectionStateConnecting
	ConnectionStateReconnecting
	ConnectionStateConnected
)
