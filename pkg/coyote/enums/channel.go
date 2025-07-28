package enums

type ChannelType int

const (
	ChannelTypeA ChannelType = iota + 1
	ChannelTypeB
)

func (c ChannelType) String() string {
	switch c {
	case ChannelTypeA:
		return "A"
	case ChannelTypeB:
		return "B"
	default:
		return "Unknown"
	}
}
