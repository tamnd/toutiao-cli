// Package cli builds the toutiao command tree on top of the toutiao-cli library.
package cli

import (
	"github.com/spf13/cobra"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Root builds the root command and its subtree.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "toutiao",
		Short: "Browse Toutiao ByteDance news feed (今日头条)",
		Long: `Browse Toutiao ByteDance news feed (今日头条)

This is a fresh scaffold. Add your commands here on top of the toutiao-cli
library package, then wire them into Root with root.AddCommand.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newVersionCmd())
	// TODO: root.AddCommand(newGetCmd()), etc.
	return root
}
