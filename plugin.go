package kindlepush

type Plugin interface {
	FetchFeed() ([]*Post, error)
}

type PluginFunc func() ([]*Post, error)

func (f PluginFunc) FetchFeed() ([]*Post, error) {
	return f()
}

type pluginEntry struct {
	id, name string
	handler  Plugin
}

var (
	allPlugins []pluginEntry
)

func RegisterPlugin(id, name string, plugin Plugin) {
	allPlugins = append(allPlugins, pluginEntry{id, name, plugin})
}
