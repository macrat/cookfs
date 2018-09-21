package plugin

type PluginPack interface {
	Name() string
	GetEndpoint(PluginType) (Plugin, error)
}

type pluginDLL struct {
	// TODO
}

func (pd pluginDLL) GetEndpoint(t PluginType) (Plugin, error) {
	// TODO
}

func loadPluginPack(name string) (pluginPack, error) {
	if bp, ok := bundledPlugins[name]; ok {
		return bp, nil
	}

	// TODO: load DLL
}

func Load(name string, t PluginType) (Plugin, error) {
	p, e := loadPluginPack(name)
	if e != nil {
		return nil, e
	}
	return p.GetEndpoint(t)
}
