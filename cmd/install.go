package cmd

import (
	"github.com/lsewcx/toolup/internal/app"
	"github.com/lsewcx/toolup/internal/execx"
	"github.com/spf13/cobra"
)

// installCmd 表示安装工具的命令。
var installCmd = &cobra.Command{
	Use:   "install <tool>",
	Short: "安装指定工具",
	Long:  `根据当前系统环境安装指定工具。`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Install(cmd.Context(), args[0], execx.ExecRunner{})
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
