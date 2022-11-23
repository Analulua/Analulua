package datadog

import (
	"newdemo1/resource/config"
	"newdemo1/resource/jaeger/common/telemetry"
)

type DataDog struct {
	metric telemetry.Metrics
	config config.Configuration
}

func NewDataDog(configuration config.Configuration, trans *telemetry.API) (DataDog, error) {
	return DataDog{
		metric: trans.Metric(),
		config: configuration,
	}, nil
}

func (d *DataDog) Metrics() telemetry.Metrics {
	return d.metric
}
