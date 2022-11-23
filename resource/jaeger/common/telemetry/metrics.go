package telemetry

// Metrics Interface. All methods SHOULD be safe for concurrent use.
type Metrics interface {
	IncrSuccess(metricName string)
	IncrFail(metricName string, err error)
	DecrSuccess(metricName string)
	DecrFail(metricName string, err error)

	IncrHTTP(method, metricName string, httpStatusCode int)
	DecrHTTP(method, metricName string, httpStatusCode int)

	Incr(metricName string, tags []string)
	Decr(metricName string, tags []string)

	Count(metricName string, value int64, tags []string)
	Gauge(metricName string, value float64, tags []string)
	Histogram(metricName string, value float64, tags []string)
	Distribution(metricName string, value float64, tags []string)
}
