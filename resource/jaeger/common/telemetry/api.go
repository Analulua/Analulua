package telemetry

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"newdemo1/resource/jaeger/common/telemetry/instrumentation/filter"
	"newdemo1/resource/jaeger/common/telemetry/internal"
)

// API is the wrapper for logger, trace, and metric instrumentation.
type API struct {
	// the instrumentation name
	name                string
	logger              Logger
	trace               trace.TracerProvider
	metric              metric.MeterProvider
	metricDD            Metrics
	MetricExportHandler http.Handler
	// Service Name For Parsing From Instrument
	ServiceAPI string
	SourceEnv  string
	Filter     *FilterConfig
}

type APIConfig struct {
	// Name is the unique instrumentationName for this instrumentation.
	Name         string
	LoggerConfig LoggerConfig
	TraceConfig  TraceConfig
	MetricConfig MetricConfig
}

type FilterConfig struct {
	PayloadFilter func(rules *filter.TargetFilter) []*regexp.Regexp
	HeaderFilter  []string
}

// LoggerConfig define the configuration logger.
type LoggerConfig struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory, if fileName is blank it will not write the log to a file.
	FileName string

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.
	MaxAge int
}

// TraceConfig define the configuration for Jaeger trace.
type TraceConfig struct {
	// CollectorEndpoint of Jaeger
	CollectorEndpoint string
	// ServiceName for jaeger filter search
	ServiceName string
	SourceEnv   string
}

// MetricConfig define the configuration for Prometheus metric.
type MetricConfig struct {
	// Port to expose the metrics
	Port         int
	AgentAddress string
	SampleRate   float64
}

const (
	name = "ndigital.telemetry"
)

// Override TraceID with xrequestid
type NewIDGen interface {
	NewTraceID() trace.ID
	NewSpanID() trace.SpanID
}
type CustomIDGen struct {
	sync.Mutex
	OtelID string
}

func (cid *CustomIDGen) NewTraceID() trace.ID {
	var arraybyte [16]byte
	copy(arraybyte[:], cid.OtelID)
	bytestr := []byte(cid.OtelID)
	uuidbyte, _ := uuid.FromBytes(bytestr)
	uuidSHA1 := uuid.NewSHA1(uuidbyte, bytestr)
	return trace.ID(uuidSHA1)
}
func (cid *CustomIDGen) NewSpanID() trace.SpanID {
	cid.Lock()
	defer cid.Unlock()
	sid := trace.SpanID{}
	_, _ = rand.Read(sid[:])
	return sid
}

func customIDGenerator(otid string) NewIDGen {
	gen := &CustomIDGen{OtelID: otid}

	return gen
}

func AddDataTracingID(otid string, oteltraceid *API) {
	provider := oteltraceid.trace.(*sdktrace.TracerProvider)
	provider.ApplyConfig(sdktrace.Config{
		IDGenerator: customIDGenerator(otid),
	})

}

func NewInstumentationWithoutMetric(config APIConfig) (*API, func(), error) {
	l, logFlushFn, err := initLogger(&config.LoggerConfig)
	if err != nil {
		return nil, nil, err
	}

	ServiceNameConfig := config.TraceConfig.ServiceName
	if config.TraceConfig.ServiceName == "" {
		ServiceNameConfig = ""
	}

	return &API{
			name:                name,
			logger:              l,
			trace:               nil,
			metric:              nil,
			MetricExportHandler: nil,
			ServiceAPI:          ServiceNameConfig,
		}, func() {
			logFlushFn()
		}, nil
}

// NewInstrumentation create and initialize all the instrumentation.
func NewNoopInstrumentation(config APIConfig) (*API, func(), error) {
	l, logFlushFn, err := initLogger(&config.LoggerConfig)
	if err != nil {
		return nil, nil, err
	}

	ServiceNameConfig := config.TraceConfig.ServiceName
	if config.TraceConfig.ServiceName == "" {
		ServiceNameConfig = ""
	}

	t, traceFlushFn, err := traceNoop()
	if err != nil {
		return nil, nil, err
	}

	m, metricFlushFn, err := metricNoop()
	if err != nil {
		return nil, nil, err
	}

	return &API{
			name:                name,
			logger:              l,
			trace:               t,
			metric:              m,
			MetricExportHandler: nil,
			ServiceAPI:          ServiceNameConfig,
		}, func() {
			logFlushFn()
			traceFlushFn()
			metricFlushFn()
		}, nil
}

// NewInstrumentation create and initialize all the instrumentation.
func NewInstrumentation(config APIConfig) (*API, func(), error) {
	l, logFlushFn, err := initLogger(&config.LoggerConfig)
	if err != nil {
		return nil, nil, err
	}

	t, traceFlushFn, err := initTrace(&config.TraceConfig)
	if err != nil {
		return nil, nil, err
	}

	m, metricFlushFn, err := initMetric(&config.MetricConfig)
	if err != nil {
		return nil, nil, err
	}

	mDD, metricDDFlushFn, err := initMetricDD(&config.MetricConfig)
	if err != nil {
		l.Error(context.Background(), "Error in Data Dog. Error: %s", err)
		return nil, nil, err
	}

	ServiceNameConfig := config.TraceConfig.ServiceName
	if config.TraceConfig.ServiceName == "" {
		ServiceNameConfig = ""
	}

	sourceEnvConfig := config.TraceConfig.SourceEnv
	if config.TraceConfig.SourceEnv == "" {
		sourceEnvConfig = ""
	}

	return &API{
			name:                name,
			logger:              l,
			trace:               t,
			metric:              m.MeterProvider(),
			metricDD:            mDD,
			MetricExportHandler: m,
			ServiceAPI:          ServiceNameConfig,
			SourceEnv:           sourceEnvConfig,
		}, func() {
			logFlushFn()
			traceFlushFn()
			metricFlushFn()
			metricDDFlushFn()
		}, nil
}

func initLogger(cfg *LoggerConfig) (Logger, func(), error) {
	c := zap.NewProductionConfig()
	c.DisableCaller = true
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	c.EncoderConfig.LevelKey = "severity"

	var options []zap.Option
	var logRotate *lumberjack.Logger
	var zp *zap.Logger
	var err error

	zp, err = c.Build(options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize zap logger %w", err)
	}

	core := zp.Core()
	if cfg.FileName != "" {
		logRotate = &lumberjack.Logger{
			Filename:  cfg.FileName,
			MaxSize:   cfg.MaxSize,
			MaxAge:    cfg.MaxAge,
			LocalTime: false,
			Compress:  true,
		}

		core = zapcore.NewTee(zp.Core(), zapcore.
			NewCore(zapcore.NewJSONEncoder(c.EncoderConfig), zapcore.AddSync(logRotate), zap.InfoLevel))

	}

	zp = zap.New(core)
	log, err := internal.NewLogger(zp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize internal logger %w", err)
	}

	return log, func() {
		_ = zp.Sync()

		if logRotate != nil {
			_ = logRotate.Close()
		}
	}, nil
}

//nolint
func traceNoop() (trace.TracerProvider, func(), error) {
	return trace.NoopTracerProvider(), func() {}, nil
}

func metricNoop() (metric.MeterProvider, func(), error) {
	return metric.NoopMeterProvider{}, func() {}, nil
}

func initTrace(config *TraceConfig) (trace.TracerProvider, func(), error) {
	tp, flushFn, err := jaeger.NewExportPipeline(
		jaeger.WithCollectorEndpoint(config.CollectorEndpoint),
		// add service name
		jaeger.WithProcess(jaeger.Process{ServiceName: config.ServiceName}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		return nil, flushFn, fmt.Errorf("failed to initialize jaeger exporter %w", err)
	}
	global.SetTracerProvider(tp)

	return tp, flushFn, nil
}

func initMetric(config *MetricConfig) (*prometheus.Exporter, func(), error) {
	exp, err := prometheus.NewExportPipeline(prometheus.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize prometheus exporter %w", err)
	}

	if config.Port == 0 {
		return exp, func() {}, nil
	}

	// Create the default mux
	mux := http.NewServeMux()
	listen := fmt.Sprintf(":%d", config.Port)
	s := &http.Server{
		Addr:    listen,
		Handler: mux,
	}

	mux.HandleFunc("/metrics", exp.ServeHTTP)

	go func() {
		_ = s.ListenAndServe()
	}()

	fmt.Println("listen on: " + listen)

	return exp, func() { _ = s.Shutdown(context.Background()) }, nil
}

func initMetricDD(config *MetricConfig) (Metrics, func(), error) {
	if config.SampleRate < 0 {
		config.SampleRate = 0
	} else if config.SampleRate > 1 {
		config.SampleRate = 1
	}

	statsd, err := statsd.New(config.AgentAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize Data dog %w", err)
	}

	m, err := internal.NewDDMetrics(statsd, config.SampleRate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize internal metrics %w", err)
	}
	// Load blocked Metric Names

	return m, func() { _ = statsd.Close() }, nil
}

func (a *API) Logger() Logger {
	return a.logger
}

func (a *API) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return a.trace.Tracer(a.name, opts...)
}

func (a *API) Meter(opts ...metric.MeterOption) metric.Meter {
	return a.metric.Meter(a.name, opts...)
}

func (a *API) Metric() Metrics {
	return a.metricDD
}
