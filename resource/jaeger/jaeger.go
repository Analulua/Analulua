package jaeger

import (
	"fmt"
	"log"
	"newdemo1/resource/config"
	Telemetry "newdemo1/resource/jaeger/common/telemetry"
	"newdemo1/resource/jaeger/common/telemetry/instrumentation/filter"
	"regexp"
)

type (
	Jaeger struct {
		Tracer  *Telemetry.API
		FlushFn func()
	}
)

func NewJaeger(config config.Configuration) (Jaeger, error) {
	api, fc, err := Telemetry.NewInstrumentation(Telemetry.APIConfig{
		LoggerConfig: Telemetry.LoggerConfig{},
		TraceConfig: Telemetry.TraceConfig{
			CollectorEndpoint: config.Telemetry.Tracer.CollectorEndpoint,
			ServiceName:       config.Telemetry.Tracer.ServiceName,
			SourceEnv:         config.Telemetry.Tracer.SourceEnv,
		},
		MetricConfig: Telemetry.MetricConfig{
			Port:         config.Telemetry.Metric.Port,
			AgentAddress: config.Telemetry.Metric.AgentAddress,
			SampleRate:   config.Telemetry.Metric.SampleRate,
		},
	})
	if err != nil {
		log.Printf("error while initializing telemetry API: %v", err)
	}
	api.Filter = &Telemetry.FilterConfig{}
	api.Filter.PayloadFilter = func(item *filter.TargetFilter) []*regexp.Regexp {
		var rules []*regexp.Regexp
		for _, v := range config.Telemetry.Filter.Body {
			pattern := fmt.Sprintf(`(%s|\"%s\"\s*):\s?"([\w\s#-@]+)`, v, v)
			regex := regexp.MustCompile(pattern)
			rules = append(rules, regex)
		}
		return rules
	}
	api.Filter.HeaderFilter = config.Telemetry.Filter.Header
	return Jaeger{
		Tracer:  api,
		FlushFn: fc,
	}, err
}
