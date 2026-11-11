package catalog

// Tool 描述一个可安装的工具。
type Tool struct {
	Name string
	// Packages 记录各系统包管理器中的包名，键为包管理器名称（apt、dnf）。
	Packages map[string]string
	// InstallScript 是工具官方安装脚本的地址。
	// 当工具的 Packages 未收录当前包管理器时（例如 Debian/Ubuntu 没有 uv），
	// 会回退到官方脚本安装，从而在不同发行版上保持一致。
	InstallScript string
	// InstallArchive 描述通过官方二进制压缩包安装的方式，
	// 适用于只提供压缩包、没有安装脚本的工具（例如 ossutil 2.0）。
	InstallArchive *Archive
}

// Archive 描述官方二进制压缩包的安装信息。
type Archive struct {
	// Version 是压缩包对应的工具版本，仅用于日志展示。
	Version string
	// Binary 是压缩包内需要安装的可执行文件名。
	Binary string
	// InstallDir 是二进制安装目录，例如 /usr/local/bin。
	InstallDir string
	// Symlink 是需要额外创建的软链接路径，例如 /usr/bin/ossutil，可为空。
	Symlink string
	// Artifacts 以 "GOOS/GOARCH" 为键（例如 linux/amd64），记录下载地址与校验和。
	Artifacts map[string]Artifact
}

// Artifact 是某个平台对应的官方下载包。
type Artifact struct {
	URL    string
	SHA256 string
}

var registry = map[string]Tool{
	"git":     gitTool,
	"uv":      uvTool,
	"ossutil": ossutilTool,
}

func Lookup(name string) (Tool, bool) {
	tool, ok := registry[name]
	return tool, ok
}
