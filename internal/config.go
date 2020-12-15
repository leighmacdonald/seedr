package internal

import (
	"github.com/leighmacdonald/seedr/pkg/client"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"sort"
)

var (
	config           configuration
	ErrInvalidConfig = errors.New("Invalid configuration")
)

type configuration struct {
	General struct {
		UpdateInterval string `mapstructure:"update_interval"`
		StatInterval   string `mapstructure:"stat_interval"`
		DryRunMode     bool   `mapstructure:"dry_run_mode"`
	} `mapstructure:"general"`
	Log struct {
		Level     string `mapstructure:"level"`
		LogColour bool   `mapstructure:"log_colour"`
	} `mapstructure:"log"`
	Client client.Config `mapstructure:"client"`
	Paths  []pathConfig  `mapstructure:"paths"`
}

type pathConfig struct {
	Path     string  `mapstructure:"path"`
	Priority int     `mapstructure:"priority"`
	MaxUsed  float64 `mapstructure:"max_used"`
	MaxRatio float64 `mapstructure:"max_ratio"`
}

// Read reads in config file and ENV variables if set.
func ReadConfig(cfgFile string) error {
	// Find home directory.
	home, _ := homedir.Dir()
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.SetConfigName("seedr")
	if os.Getenv("SEEDR_CONFIG") != "" {
		viper.SetConfigFile(os.Getenv("SEEDR_CONFIG"))
	} else if cfgFile != "" {
		viper.SetConfigName(cfgFile)
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
		newConfig := configuration{}
		if err := viper.Unmarshal(&newConfig); err != nil {
			return errors.Wrapf(err, "Failed to parse config")
		}
		config = newConfig

		setupLogger(config.Log.Level, config.Log.LogColour)
		return nil
	}
	return ErrInvalidConfig
}

func setupLogger(levelStr string, colour bool) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      colour,
		DisableTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(levelStr)
	if err != nil {
		log.Panicln("Invalid log level defined")
	}
	log.SetLevel(level)
}

func Save() error {
	return viper.WriteConfig()
}

func configSanityCheck() {

}

func pathsByPriority() []pathConfig {
	paths := config.Paths
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Priority > paths[j].Priority
	})
	return paths
}
