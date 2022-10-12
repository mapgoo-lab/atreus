package orm

import "github.com/mapgoo-lab/atreus/pkg/stat/metric"

const namespace = "orm"

var (
	_metricOrmDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "gorm requests duration(ms).",
		Labels:    []string{"table", "command"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
	})
	_metricOrmErr = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: namespace,
		Subsystem: "requests",
		Name:      "error_total",
		Help:      "gorm requests error count.",
		Labels:    []string{"table", "command", "error"},
	})
)