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

package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/deflinhec/rfslog/internal"
	"github.com/dietsche/rfsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v3"
)

var (
	Version = "0.0.0"
	Build   = "-"
)

func main() {
	semver := fmt.Sprintf("%s+%s", Version, Build)

	ctx, ctxCancelFn := context.WithCancel(context.Background())
	defer ctxCancelFn()

	tmpLogger := internal.NewJSONLogger(os.Stdout, zapcore.InfoLevel, internal.JSONFormat)
	tmpLogger.Sync()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println(semver)
			os.Exit(0)
		case "configfile":
			config := internal.NewConfig(tmpLogger)
			data, _ := yaml.Marshal(config)
			os.WriteFile("config.yaml", data, 0644)
			os.Exit(0)
		}
	}

	config := internal.ParseArgs(tmpLogger, os.Args)
	logger, startupLogger := internal.SetupLogging(tmpLogger, config)
	defer logger.Sync()

	logger = logger.WithOptions(zap.WithCaller(false))
	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		startupLogger.Fatal("Failed to create runtime directory watcher", zap.Error(err))
	}
	if err = watcher.AddRecursive(config.GetWatch()); err != nil {
		startupLogger.Fatal("An error occurred while watching directory", zap.Error(err))
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Context cancelled
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				var checksum string
				switch event.Op {
				case fsnotify.Remove:
				case fsnotify.Rename:
				default:
					if b, err := os.ReadFile(event.Name); err != nil {
						logger.Error("Failed to read file",
							zap.String("event", event.Op.String()),
							zap.String("file", event.Name),
							zap.Error(err),
						)
					} else {
						hash := md5.Sum(b)
						checksum = hex.EncodeToString(hash[:])
					}
				}
				logger.Info("Detected",
					zap.String("event", event.Op.String()),
					zap.String("file", event.Name),
					zap.String("md5sum", checksum),
				)
			}
		}
	}()

	startupLogger.Info("Starting rfslog")
	startupLogger.Info("Process", zap.String("version", semver), zap.String("runtime", runtime.Version()))
	startupLogger.Info("Watching directory", zap.String("directory", config.GetWatch()))

	// Respect OS stop signals.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-c

	logger.Info("Shutting down")
}
