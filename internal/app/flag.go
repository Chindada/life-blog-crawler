package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	urlHome     string = "https://tocandraw.com/"
	urlPost     string = "https://tocandraw.com/post-sitemap.xml"
	urlPage     string = "https://tocandraw.com/page-sitemap.xml"
	urlCategory string = "https://tocandraw.com/category-sitemap.xml"
	urlTag      string = "https://tocandraw.com/post_tag-sitemap.xml"
	urlAuthor   string = "https://tocandraw.com/author-sitemap.xml"
)

func ReadCommand() error {
	rootCmd := &cobra.Command{
		Use:          "lbcrawler",
		Version:      "v0.0.1",
		Short:        "life blog utility",
		Long:         "It is a life blog utility",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
	}

	rootCmd.AddCommand(RunCommand())
	return rootCmd.Execute()
}

func RunCommand() *cobra.Command {
	allSitemap := []string{urlPost, urlPage, urlCategory, urlTag, urlAuthor}
	c := &cobra.Command{
		Use:   "run",
		Short: "run crawler",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if auth := viper.GetString("cloudflare-auth"); auth == "" {
				return fmt.Errorf("cloudflare auth key is required")
			}
			if zone := viper.GetString("cloudflare-zone"); zone == "" {
				return fmt.Errorf("cloudflare zone is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			crawler, err := NewCrawler(allSitemap, false, urlHome)
			if err != nil {
				return err
			}
			if err = crawler.PurgeCache(); err != nil {
				return err
			}
			crawler.Run()

			mobileCrawler, err := NewCrawler(allSitemap, true, urlHome)
			if err != nil {
				return err
			}
			mobileCrawler.Run()
			return nil
		},
	}

	c.Flags().String("cloudflare-auth", "", "cloudflare auth key")
	_ = viper.BindPFlag("cloudflare-auth", c.Flags().Lookup("cloudflare-auth"))
	c.Flags().String("cloudflare-zone", "", "cloudflare zone")
	_ = viper.BindPFlag("cloudflare-zone", c.Flags().Lookup("cloudflare-zone"))
	return c
}
