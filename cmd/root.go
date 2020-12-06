package cmd

import (
	"github.com/leighmacdonald/seedr/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "hugo",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.Start()
	},
}

func Execute() {
	if err := internal.ReadConfig(""); err != nil {
		log.Errorf("Failed to read config: %v", err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
