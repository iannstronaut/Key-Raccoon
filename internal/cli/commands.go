package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"keyraccoon/internal/config"
)

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "keyraccoon",
		Short: "KeyRaccoon CLI",
	}

	rootCmd.AddCommand(newConfigCommand())
	return rootCmd
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print effective runtime config summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Init()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "environment=%s\nserver=%s:%s\ndatabase=%s:%s/%s\nredis=%s:%s\n",
				cfg.Environment,
				cfg.ServerHost,
				cfg.ServerPort,
				cfg.DBHost,
				cfg.DBPort,
				cfg.DBName,
				cfg.RedisHost,
				cfg.RedisPort,
			)
			return nil
		},
	}
}
