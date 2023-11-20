// Copyright The OpenTelemetry Authors
// Copyright (c) 2018 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package main // import "github.com/open-telemetry/opentelemetry-collector-contrib/telemetrygen/internal/telemetrygen"

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	file_cmd "github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/config_file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/logs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/metrics"
	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/traces"
)

var (
	tracesCfg   *traces.Config
	metricsCfg  *metrics.Config
	logsCfg     *logs.Config
	filePathCfg *file_cmd.FileCmdConfig
)

// rootCmd is the root command on which will be run children commands
var rootCmd = &cobra.Command{
	Use:     "telemetrygen",
	Short:   "Telemetrygen simulates a client generating traces, metrics, and logs",
	Example: "telemetrygen traces\ntelemetrygen metrics\ntelemetrygen logs",
}

// tracesCmd is the command responsible for sending traces
var tracesCmd = &cobra.Command{
	Use:     "traces",
	Short:   fmt.Sprintf("Simulates a client generating traces. (Stability level: %s)", metadata.TracesStability),
	Example: "telemetrygen traces",
	RunE: func(cmd *cobra.Command, args []string) error {
		return traces.Start(tracesCfg)
	},
}

// metricsCmd is the command responsible for sending metrics
var metricsCmd = &cobra.Command{
	Use:     "metrics",
	Short:   fmt.Sprintf("Simulates a client generating metrics. (Stability level: %s)", metadata.MetricsStability),
	Example: "telemetrygen metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return metrics.Start(metricsCfg)
	},
}

// logsCmd is the command responsible for sending logs
var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   fmt.Sprintf("Simulates a client generating logs. (Stability level: %s)", metadata.LogsStability),
	Example: "telemetrygen logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return logs.Start(logsCfg)
	},
}

var fileCmd = &cobra.Command{
	Use:     "file",
	Short:   "Provide a yml config file to generate multiple telemtry streams",
	Example: "telemetrygen config=config.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		return file_cmd.Start(filePathCfg)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd, tracesCmd, metricsCmd, logsCmd)

	tracesCfg = new(traces.Config)
	tracesCfg.Flags(tracesCmd.Flags())

	metricsCfg = new(metrics.Config)
	metricsCfg.Flags(metricsCmd.Flags())

	logsCfg = new(logs.Config)
	logsCfg.Flags(logsCmd.Flags())

	filePathCfg = new(file_cmd.FileCmdConfig)
	filePathCfg.Flags(fileCmd.Flags())

	// Disabling completion command for end user
	// https://github.com/spf13/cobra/blob/master/shell_completions.md
	rootCmd.CompletionOptions.DisableDefaultCmd = true

}

// Execute tries to run the input command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// TODO: Uncomment the line below when using Run instead of RunE in the xxxCmd functions
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
