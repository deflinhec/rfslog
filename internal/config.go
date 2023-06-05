// Copyright 2023 Deflihnec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"flag"
	"os"

	"github.com/heroiclabs/nakama/v3/flags"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config interface {
	GetWatch() string
	GetLogger() *LoggerConfig

	Clone() (Config, error)
}

func ParseArgs(logger *zap.Logger, args []string) Config {
	// Parse args to get path to a config file if passed in.
	configFilePath := NewConfig(logger)
	configFileFlagSet := flag.NewFlagSet("rfslog", flag.ExitOnError)
	configFileFlagMaker := flags.NewFlagMakerFlagSet(&flags.FlagMakingOptions{
		UseLowerCase: true,
		Flatten:      false,
		TagName:      "yaml",
		TagUsage:     "usage",
	}, configFileFlagSet)

	if _, err := configFileFlagMaker.ParseArgs(configFilePath, args[1:]); err != nil {
		logger.Fatal("Could not parse command line arguments", zap.Error(err))
	}

	// Parse config file if path is set.
	mainConfig := NewConfig(logger)
	for _, cfg := range configFilePath.Config {
		data, err := os.ReadFile(cfg)
		if err != nil {
			logger.Fatal("Could not read config file", zap.String("path", cfg), zap.Error(err))
		}

		err = yaml.Unmarshal(data, mainConfig)
		if err != nil {
			logger.Fatal("Could not parse config file", zap.String("path", cfg), zap.Error(err))
		}
	}
	// Preserve the config file path arguments.
	mainConfig.Config = configFilePath.Config

	// Override config with those passed from command-line.
	mainFlagSet := flag.NewFlagSet("rfslog", flag.ExitOnError)
	mainFlagMaker := flags.NewFlagMakerFlagSet(&flags.FlagMakingOptions{
		UseLowerCase: true,
		Flatten:      false,
		TagName:      "yaml",
		TagUsage:     "usage",
	}, mainFlagSet)

	if _, err := mainFlagMaker.ParseArgs(mainConfig, args[1:]); err != nil {
		logger.Fatal("Could not parse command line arguments", zap.Error(err))
	}

	return mainConfig
}

type config struct {
	Watch  string        `yaml:"watch" json:"watch" usage:"The directory to watch."`
	Config []string      `yaml:"config" json:"config" usage:"The absolute file path to configuration YAML file."`
	Logger *LoggerConfig `yaml:"logger" json:"logger" usage:"Logger levels and output."`
}

// NewConfig constructs a Config struct which represents server settings, and populates it with default values.
func NewConfig(logger *zap.Logger) *config {
	return &config{
		Watch:  ".",
		Logger: NewLoggerConfig(),
	}
}

func (c *config) Clone() (Config, error) {
	configLogger := *(c.Logger)
	nc := &config{
		Watch:  c.Watch,
		Logger: &configLogger,
	}
	return nc, nil
}

func (c *config) GetWatch() string {
	return c.Watch
}

func (c *config) GetLogger() *LoggerConfig {
	return c.Logger
}

// LoggerConfig is configuration relevant to logging levels and output.
type LoggerConfig struct {
	Level    string `yaml:"level" json:"level" usage:"Log level to set. Valid values are 'debug', 'info', 'warn', 'error'. Default 'info'."`
	Stdout   bool   `yaml:"stdout" json:"stdout" usage:"Log to standard console output (as well as to a file if set). Default true."`
	File     string `yaml:"file" json:"file" usage:"Log output to a file (as well as stdout if set). Make sure that the directory and the file is writable."`
	Rotation bool   `yaml:"rotation" json:"rotation" usage:"Rotate log files. Default is false."`
	// Reference: https://godoc.org/gopkg.in/natefinch/lumberjack.v2
	MaxSize    int    `yaml:"max_size" json:"max_size" usage:"The maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes."`
	MaxAge     int    `yaml:"max_age" json:"max_age" usage:"The maximum number of days to retain old log files based on the timestamp encoded in their filename. The default is not to remove old log files based on age."`
	MaxBackups int    `yaml:"max_backups" json:"max_backups" usage:"The maximum number of old log files to retain. The default is to retain all old log files (though MaxAge may still cause them to get deleted.)"`
	LocalTime  bool   `yaml:"local_time" json:"local_time" usage:"This determines if the time used for formatting the timestamps in backup files is the computer's local time. The default is to use UTC time."`
	Compress   bool   `yaml:"compress" json:"compress" usage:"This determines if the rotated log files should be compressed using gzip."`
	Format     string `yaml:"format" json:"format" usage:"Set logging output format. Can either be 'JSON' or 'Stackdriver'. Default is 'JSON'."`
}

func NewLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "info",
		Stdout:     true,
		File:       "",
		Rotation:   false,
		MaxSize:    100,
		MaxAge:     0,
		MaxBackups: 0,
		LocalTime:  false,
		Compress:   false,
		Format:     "json",
	}
}
