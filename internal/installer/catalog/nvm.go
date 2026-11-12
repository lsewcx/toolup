package catalog

// nvmTool 使用 nvm 官方安装脚本，并在每次安装时解析 GitHub 最新 release。
//
// 文档：https://github.com/nvm-sh/nvm#install--update-script
// 官方命令示例：curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.41.0/install.sh | bash
//
// 这里把 URL 写成带 {version} 占位符的形式，安装前先请求 GitHub 的
// releases/latest 重定向地址解析出最新 tag，从而始终安装最新版本。
//
// 脚本会把 nvm 安装到 ~/.nvm（设置了 XDG_CONFIG_HOME 时为 $XDG_CONFIG_HOME/nvm），
// 并把加载 nvm 的配置写入 shell 配置文件。nvm 是 shell 函数而不是可执行文件，
// 因此不能通过 PATH 检测，改为判断 nvm.sh 是否存在。
var nvmTool = Tool{
	Name:                   "nvm",
	InstallScript:          "https://raw.githubusercontent.com/nvm-sh/nvm/{version}/install.sh",
	InstallScriptLatestURL: "https://github.com/nvm-sh/nvm/releases/latest",
	InstallScriptShell:     "bash",
	InstallScriptInstalledPaths: []string{
		"$HOME/.nvm/nvm.sh",
		"$XDG_CONFIG_HOME/nvm/nvm.sh",
	},
}
