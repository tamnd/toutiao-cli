package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) feedCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "List Toutiao news feed articles",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			articles, err := a.client.Feed(cmd.Context(), category, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}
	cmd.Flags().StringVarP(&category, "category", "c", "hot", "category: hot|video|all")
	return cmd
}

func (a *App) hotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hot",
		Short: "List Toutiao hot news (今日头条热点)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			articles, err := a.client.Feed(cmd.Context(), "hot", n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(articles, len(articles))
		},
	}
}
