// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"github.com/spf13/pflag"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen/internal/common"
)

// Config describes the test scenario.
type Config struct {
	common.Config
	NumMetrics            int
	MetricType            metricType
	UseRandomValues       bool
	HistogramBucketBounds []float64
}

// Flags registers config flags.
func (c *Config) Flags(fs *pflag.FlagSet) {
	// Use Gauge as default metric type.
	c.MetricType = metricTypeGauge

	c.CommonFlags(fs)

	fs.StringVar(&c.HTTPPath, "otlp-http-url-path", "/v1/metrics", "Which URL path to write to")

	fs.Var(&c.MetricType, "metric-type", "Metric type enum. must be one of 'Gauge' or 'Sum'")
	fs.IntVar(&c.NumMetrics, "metrics", 1, "Number of metrics to generate in each worker (ignored if duration is provided)")

	fs.BoolVar(&c.UseRandomValues, "random-metric-values", false, "Use random values for metrics")

	fs.Float64SliceVar(&c.HistogramBucketBounds, "histogram-bucket-bounds", []float64{2, 4, 8, 16, 32, 64, 128}, "For histogram metrics, defines the bounds of each bucket")
}
