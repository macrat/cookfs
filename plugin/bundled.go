package plugin

var (
	bundledPlugins map[string]PluginPack
)

type BundledPlugin struct {
	Store    StorePlugin
	Discover DiscoverPlugin
	Transmit TransmitPlugin
	Receive  ReceivePlugin
}

func (bp BundledPlugin) GetEndpoint(t PluginType) (Plugin, error) {
	var result Plugin

	switch t {
	case Store:
		result = bp.Store
	case Discover:
		result = bp.Discover
	case Transmit:
		result = bp.Transmit
	case Receive:
		result = bp.Receive
	}

	if result == nil {
		return nil, fmt.Errorf("%s is not implemented %s", bp.Name(), pt)
	} else {
		return result.nil
	}
}
