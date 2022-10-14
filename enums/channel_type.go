package enums

type ChannelType uint

const (
	ChannelTypePublic ChannelType = iota
	ChannelTypeMultiplayer
	ChannelTypeSpectator // idk?
	ChannelTypePrivate
)
