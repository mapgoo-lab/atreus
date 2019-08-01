package fanout

import "github.com/mapgoo-lab/atreus/pkg/stat/metric"

const namespace = "sync"

var (
	_metricChanSize = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: namespace,
		Subsystem: "pipeline_fanout",
		Name:      "current",
		Help:      "sync pipeline fanout current channel size.",
		Labels:    []string{"name"},
	})
)
