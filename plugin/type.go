package plugin

type PluginType int

var (
	Unknown PluginType = iota
	Store
	Discover
	Transmit
	Receive
)

func (pt PluginType) String() string {
	switch pt {
	case Unknown:
		return "UnknownPlugin"
	case Store:
		return "StorePlugin"
	case Discover:
		return "DiscoverPlugin"
	case Transmit:
		return "TransmitPlugin"
	case Receive:
		return "ReceivePlugin"
	}
}
