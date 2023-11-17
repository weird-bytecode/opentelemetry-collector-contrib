// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type worker struct {
	running               *atomic.Bool // pointer to shared flag that indicates it's time to stop the test
	metricType            metricType   // type of metric to generate
	numMetrics            int          // how many metrics the worker has to generate (only when duration==0)
	useRandomValues       bool
	histogramBucketBounds []float64
	totalDuration         time.Duration   // how long to run the test for (overrides `numMetrics`)
	limitPerSecond        rate.Limit      // how many metrics per second to generate
	wg                    *sync.WaitGroup // notify when done
	logger                *zap.Logger     // logger
	index                 int             // worker index
}

func (w worker) simulateMetrics(res *resource.Resource, exporterFunc func() (sdkmetric.Exporter, error), signalAttrs []attribute.KeyValue) {
	limiter := rate.NewLimiter(w.limitPerSecond, 1)

	exporter, err := exporterFunc()
	if err != nil {
		w.logger.Error("failed to create the exporter", zap.Error(err))
		return
	}

	defer func() {
		w.logger.Info("stopping the exporter")
		if tempError := exporter.Shutdown(context.Background()); tempError != nil {
			w.logger.Error("failed to stop the exporter", zap.Error(tempError))
		}
	}()
	var i int64
	histogramBucketBounds := w.histogramBucketBounds
	histogramUpmostBound := int64(histogramBucketBounds[len(histogramBucketBounds)-1])
	for w.running.Load() {
		var metrics []metricdata.Metrics

		// TODO: extract functions
		switch w.metricType {
		case metricTypeGauge:
			var value int64 = i
			if w.useRandomValues {
				value = rand.Int63n(10000)
			}
			metrics = append(metrics, metricdata.Metrics{
				Name: "gen",
				Data: metricdata.Gauge[int64]{
					DataPoints: []metricdata.DataPoint[int64]{
						{
							Time:       time.Now(),
							Value:      value,
							Attributes: attribute.NewSet(signalAttrs...),
						},
					},
				},
			})
		case metricTypeSum:
			var value int64 = i
			var isMonotonic bool = true
			if w.useRandomValues {
				value = rand.Int63n(10000)
				isMonotonic = false
			}
			metrics = append(metrics, metricdata.Metrics{
				Name: "gen",
				Data: metricdata.Sum[int64]{
					IsMonotonic: isMonotonic,
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.DataPoint[int64]{
						{
							StartTime:  time.Now().Add(-1 * time.Second),
							Time:       time.Now(),
							Value:      value,
							Attributes: attribute.NewSet(signalAttrs...),
						},
					},
				},
			})
		case metricTypeHistogram:
			var count = 1
			var sum = int64(i)
			histogramBucketCounts := make([]uint64, len(histogramBucketBounds)+1)
			for i := range histogramBucketCounts {
				histogramBucketCounts[i] = 0
			}
			if w.useRandomValues {
				count = rand.Intn(100) + 1
				sum = int64(0)
				for j := 0; j < count; j++ {
					randomValue := rand.Int63n(histogramUpmostBound)
					for k, max := range histogramBucketBounds {
						if randomValue < int64(max) {
							sum += randomValue
							histogramBucketCounts[k]++
							break
						}
					}
				}
			} else {
				if i > histogramUpmostBound {
					histogramBucketCounts[len(histogramBucketCounts)-1]++
				} else {
					for k, max := range histogramBucketBounds {
						if i < int64(max) {
							histogramBucketCounts[k]++
							break
						}
					}
				}
			}
			metrics = append(metrics, metricdata.Metrics{
				Name: "gen",
				Data: metricdata.Histogram[int64]{
					Temporality: metricdata.CumulativeTemporality,
					DataPoints: []metricdata.HistogramDataPoint[int64]{
						{
							StartTime:    time.Now().Add(-1 * time.Second),
							Time:         time.Now(),
							Attributes:   attribute.NewSet(signalAttrs...),
							Sum:          sum,
							Count:        uint64(count),
							Bounds:       histogramBucketBounds,
							BucketCounts: histogramBucketCounts,
						},
					},
				},
			})
		default:
			w.logger.Fatal("unknown metric type")
		}

		rm := metricdata.ResourceMetrics{
			Resource:     res,
			ScopeMetrics: []metricdata.ScopeMetrics{{Metrics: metrics}},
		}

		if err := exporter.Export(context.Background(), &rm); err != nil {
			w.logger.Fatal("exporter failed", zap.Error(err))
		}
		if err := limiter.Wait(context.Background()); err != nil {
			w.logger.Fatal("limiter wait failed, retry", zap.Error(err))
		}

		i++
		if w.numMetrics != 0 && i >= int64(w.numMetrics) {
			break
		}
	}

	w.logger.Info("metrics generated", zap.Int64("metrics", i))
	w.wg.Done()
}
