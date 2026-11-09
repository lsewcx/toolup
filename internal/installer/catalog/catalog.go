package catalog

type Tool struct {
	Name     string
	Packages map[string]string
}

var registry = map[string]Tool{
	"git": gitTool,
}

func Lookup(name string) (Tool, bool) {
	tool, ok := registry[name]
	return tool, ok
}
