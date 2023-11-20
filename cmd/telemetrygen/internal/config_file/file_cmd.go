package file_cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/logs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/metrics"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/traces"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	fileConfig *FileConfig
)

type FileCmdConfig struct {
	configPath string
}

func (c *FileCmdConfig) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&c.configPath, "config-file", "", "The relative path to a configuration file")
}

type FileConfig struct {
	Traces  []traces.Config  `yaml:"traces,flow"`
	Metrics []metrics.Config `yaml:"metrics,flow"`
	Logs    []logs.Config    `yaml:"logs,flow"`
}

func Start(fileCmdConfig *FileCmdConfig) error {
	if len(fileCmdConfig.configPath) == 0 {
		fmt.Printf("No config file path provided, exiting...")
		return nil
	}

	err := readConfigFile(fileCmdConfig.configPath)
	if err != nil {
		fmt.Printf("Error loading file %q: %w", fileCmdConfig.configPath, err)
		return err
	}

	var wg sync.WaitGroup

	if len(fileConfig.Metrics) > 0 {
		wg.Add(len(fileConfig.Metrics))
		for _, metricsConfig := range fileConfig.Metrics {
			fmt.Printf("metricsConfig %v", metricsConfig)
			config := metricsConfig
			go func() {
				metrics.Start(&config)
				wg.Done()
			}()
		}
	}

	if len(fileConfig.Logs) > 0 {
		wg.Add(len(fileConfig.Logs))
		for _, logsConfig := range fileConfig.Logs {
			fmt.Printf("logsConfig %v", logsConfig)
			config := logsConfig
			go func() {
				logs.Start(&config)
				wg.Done()
			}()
		}
	}

	if len(fileConfig.Traces) > 0 {
		wg.Add(len(fileConfig.Traces))
		for _, tracesConfig := range fileConfig.Traces {
			fmt.Printf("tracesConfig %v", tracesConfig)
			config := tracesConfig
			go func() {
				traces.Start(&config)
				wg.Done()
			}()
		}
	}
	wg.Wait()

	return nil
}

func readConfigFile(path string) error {
	fileBuffer, fileReadErr := os.ReadFile(path)
	if fileReadErr != nil {
		return fileReadErr
	}

	fileConfig = new(FileConfig)
	unmarshalErr := yaml.Unmarshal(fileBuffer, fileConfig)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	return nil
}
