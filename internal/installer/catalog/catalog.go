package catalog

type Tool struct {
	Name string
	// Packages 记录各系统包管理器中的包名，键为包管理器名称（apt、dnf）。
	Packages map[string]string
	// InstallScript 是工具官方安装脚本的地址。
	// 当工具的 Packages 未收录当前包管理器时（例如 Debian/Ubuntu 没有 uv），
	// 会回退到官方脚本安装，从而在不同发行版上保持一致。
	InstallScript string
}

var registry = map[string]Tool{
	"git": gitTool,
	"uv":  uvTool,
}

func Lookup(name string) (Tool, bool) {
	tool, ok := registry[name]
	return tool, ok
}
