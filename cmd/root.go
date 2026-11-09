package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 是 toolup 的根命令，负责承载所有子命令。
var rootCmd = &cobra.Command{
	Use:   "toolup",
	Short: "快速安装服务器常用工具",
	Long: `toolup 用于在新服务器上快速安装所需工具。

它会自动识别当前环境，并为不同平台提供一致的安装体验。`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute 执行根命令，由 main 函数调用。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
