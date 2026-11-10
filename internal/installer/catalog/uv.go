package catalog

// uvTool 使用 Astral 官方的独立安装脚本。
// uv 没有被 Debian/Ubuntu 等发行版的仓库收录，官方脚本是跨发行版最可靠的安装方式。
var uvTool = Tool{
	Name:          "uv",
	InstallScript: "https://astral.sh/uv/install.sh",
}
