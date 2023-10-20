package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ReadCommand() {
	rootCmd := &cobra.Command{
		Use:     "lbcrawler",
		Version: "v0.0.1",
		Short:   "life blog utility",
		Long:    "It is a life blog utility",
		RunE: func(cmd *cobra.Command, args []string) error {
			if auth := viper.GetString("cloudflare-auth"); auth == "" {
				return fmt.Errorf("cloudflare auth key is required")
			}
			if zone := viper.GetString("cloudflare-zone"); zone == "" {
				return fmt.Errorf("cloudflare zone is required")
			}
			return nil
		},
	}

	rootCmd.Flags().String("cloudflare-auth", "", "cloudflare auth key")
	_ = viper.BindPFlag("cloudflare-auth", rootCmd.Flags().Lookup("cloudflare-auth"))
	rootCmd.Flags().String("cloudflare-zone", "", "cloudflare zone")
	_ = viper.BindPFlag("cloudflare-zone", rootCmd.Flags().Lookup("cloudflare-zone"))

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
