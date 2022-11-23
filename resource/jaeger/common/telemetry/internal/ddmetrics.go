package internal

import (
	"errors"
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"newdemo1/resource/jaeger/common/telemetry/instrumentation"
)

type DDMetrics struct {
	client     *statsd.Client
	sampleRate float64
}

var (
	ErrDataDogRequired = errors.New("data dog required")
)

// NewDDMetrics create new instant for the metrics.
func NewDDMetrics(c *statsd.Client, sampleRate float64) (*DDMetrics, error) {
	if c == nil {
		return nil, ErrDataDogRequired
	}
	return &DDMetrics{
		client:     c,
		sampleRate: sampleRate,
	}, nil
}

func (d *DDMetrics) Incr(metricName string, tags []string) {
	_ = d.client.Incr(metricName, tags, d.sampleRate)
}

func (d *DDMetrics) IncrSuccess(metricName string) {
	d.Incr(metricName, []string{instrumentation.MetricResponseCodeSuccess})
}
func (d *DDMetrics) IncrFail(metricName string, err error) {
	d.Incr(metricName, []string{instrumentation.GetResponseCode(err)})
}
func (d *DDMetrics) IncrHTTP(method, metricName string, httpStatusCode int) {
	// MetricName Syntax:{HTTPMethod}_{URL}
	m := fmt.Sprintf("%s_%s", method, metricName)
	rsCode := fmt.Sprintf("%s:%d", instrumentation.MetricResponseCode, httpStatusCode)
	d.Incr(m, []string{rsCode})
}

func (d *DDMetrics) DecrHTTP(method, metricName string, httpStatusCode int) {
	// MetricName Syntax:{HTTPMethod}_{URL}
	m := fmt.Sprintf("%s_%s", method, metricName)
	rsCode := fmt.Sprintf("%s:%d", instrumentation.MetricResponseCode, httpStatusCode)
	d.Decr(m, []string{rsCode})
}

func (d *DDMetrics) Decr(metricName string, tags []string) {
	_ = d.client.Decr(metricName, tags, d.sampleRate)
}

func (d *DDMetrics) DecrSuccess(metricName string) {
	d.Decr(metricName, []string{instrumentation.MetricResponseCodeSuccess})
}
func (d *DDMetrics) DecrFail(metricName string, err error) {
	d.Decr(metricName, []string{instrumentation.GetResponseCode(err)})
}

func (d *DDMetrics) Count(metricName string, value int64, tags []string) {
	_ = d.client.Count(metricName, value, tags, d.sampleRate)
}

func (d *DDMetrics) Gauge(metricName string, value float64, tags []string) {
	_ = d.client.Gauge(metricName, value, tags, d.sampleRate)
}

func (d *DDMetrics) Histogram(metricName string, value float64, tags []string) {
	_ = d.client.Histogram(metricName, value, tags, d.sampleRate)
}

func (d *DDMetrics) Distribution(metricName string, value float64, tags []string) {
	_ = d.client.Distribution(metricName, value, tags, d.sampleRate)
}
