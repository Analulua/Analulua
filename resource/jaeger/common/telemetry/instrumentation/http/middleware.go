package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/negroni"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/semconv"
	"go.uber.org/zap"

	"newdemo1/resource/jaeger/common/telemetry"
	ins "newdemo1/resource/jaeger/common/telemetry/instrumentation"
	"newdemo1/resource/jaeger/common/telemetry/instrumentation/filter"
)

const (
	CurrentRouterTemplate = "Current-Router-Template"
)

var propagator = otel.NewCompositeTextMapPropagator(propagators.TraceContext{},
	b3.B3{InjectEncoding: b3.B3MultipleHeader | b3.B3SingleHeader})
var optionshttp = []otelhttptrace.Option{otelhttptrace.WithPropagators(propagator)}

func Telemetry(handler http.Handler, api *telemetry.API) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add custom and integrate tracer span with new instrumentation
		ctx := r.Context()
		path := r.URL.Path

		tracer := api.Tracer(trace.WithInstrumentationVersion(otelcontrib.SemVersion()))
		_, _, spanCtx := otelhttptrace.Extract(ctx, r, optionshttp...)
		ctx, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), r.URL.Path)
		defer span.End()

		span.SetAttributes(ins.Method.String(r.Method+" "+r.URL.Path),
			ins.CorrelationID.String(span.SpanContext().TraceID.String()+"-"+span.SpanContext().SpanID.String()),
			semconv.HTTPClientIPKey.String(r.Header.Get(ins.XForwardedFor)),
			semconv.HTTPUserAgentKey.String(r.Header.Get(ins.UserAgent)))

		logRequest(ctx, r, api)

		// add new response writer for status
		lrw := negroni.NewResponseWriter(w)

		var appResponseCode string

		defer func(start time.Time, rw negroni.ResponseWriter) {
			elapsedTime := time.Since(start).Milliseconds()

			// send metrics to datadog
			method := strings.ToLower(r.Method)
			name := fmt.Sprintf("%s_%s", method, r.URL.Path)
			if api.ServiceAPI != "" {
				name = api.ServiceAPI
			}
			src_env := api.SourceEnv

			urlPath := r.URL.Path
			currentRouterTemplate := rw.Header().Get(CurrentRouterTemplate)
			if currentRouterTemplate != "" {
				urlPath = rw.Header().Get(CurrentRouterTemplate)
				rw.Header().Del(CurrentRouterTemplate)
			}

			endpoint := fmt.Sprintf("%s.%s", urlPath, method)

			// send metrics to datadog
			tags := []string{
				"http_endpoint:" + endpoint,
				"src_env:" + src_env,
				"response_code:" + strconv.FormatUint(uint64(rw.Status()), 10),
				"app_response_code:" + appResponseCode,
			}

			api.Metric().Count(name, 1, tags)
			api.Metric().Histogram(name+".histogram", float64(elapsedTime), tags)
			api.Metric().Distribution(name+".distribution", float64(elapsedTime), tags)

			// send metrics to open telemetry
			if vr, err := api.Meter().NewInt64ValueRecorder(path + ".valuerecorder"); err == nil {
				vr.Record(ctx, elapsedTime, label.String(ins.LabelMethod, r.Method+" "+path),
					label.Int(ins.LabelStatus, rw.Status()))
			}
		}(time.Now(), lrw)

		respLogger := &httpResponseLogger{
			context:         ctx,
			logger:          api.Logger(),
			writer:          lrw,
			httpRequest:     r,
			api:             api,
			appResponseCode: &appResponseCode,
		}

		r = r.WithContext(ctx)
		handler.ServeHTTP(respLogger, r)
	})
}

func logRequest(ctx context.Context, r *http.Request, api *telemetry.API) {
	path := r.URL.Path
	header := r.Header
	if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut {
		var requestBody string
		rawBody, err := ioutil.ReadAll(r.Body)
		if err == nil {
			r.Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))
			requestBody = string(rawBody)
		}

		var filteredBody interface{}
		headerKeyFilter := filter.DefaultHeaderFilter
		// set request and response default
		filteredBody = requestBody
		if api.Filter != nil {
			rules := api.Filter.PayloadFilter(&filter.TargetFilter{
				Method: path,
			})

			filteredBody = filter.BodyFilter(rules, requestBody)
			headerKeyFilter = append(headerKeyFilter, api.Filter.HeaderFilter...)
		}

		filteredHeader := filter.HeaderFilter(header, headerKeyFilter)

		api.Logger().Info(ctx, "Http Request",
			zap.String(ins.LabelHTTPService, path),
			zap.Any(ins.LabelHTTPHeader, filteredHeader),
			zap.Any(ins.LabelHTTPRequest, filteredBody),
			zap.Any(ins.LabelHTTPMethod, r.Method),
		)

	} else {
		headerKeyFilter := filter.DefaultHeaderFilter
		if api.Filter != nil {
			headerKeyFilter = append(headerKeyFilter, api.Filter.HeaderFilter...)
		}

		filteredHeader := filter.HeaderFilter(header, headerKeyFilter)

		api.Logger().Info(ctx, "Http Request",
			zap.String(ins.LabelHTTPService, path),
			zap.Any(ins.LabelHTTPHeader, filteredHeader),
			zap.Any(ins.LabelHTTPMethod, r.Method),
		)
	}
}

type httpResponseLogger struct {
	context         context.Context
	logger          telemetry.Logger
	writer          negroni.ResponseWriter
	httpRequest     *http.Request
	api             *telemetry.API
	appResponseCode *string
}

type httpResponseBody struct {
	ResponseCode string `json:"responseCode"`
}

func (hrl *httpResponseLogger) Header() http.Header {
	return hrl.writer.Header()
}

func (hrl *httpResponseLogger) Write(bytes []byte) (int, error) {
	// add metric component
	path := hrl.httpRequest.URL.Path
	n, err := hrl.writer.Write(bytes)
	if err != nil {
		return 0, err
	}

	filterConfig := hrl.api.Filter
	// filter response without credential field
	var makemapresp interface{}

	makemapresp = string(bytes)
	if string(bytes) != "" && filterConfig != nil {
		rules := filterConfig.PayloadFilter(&filter.TargetFilter{
			Method: path,
		})

		makemapresp = filter.BodyFilter(rules, makemapresp)
	}

	// setting up response code
	var respBody httpResponseBody
	err = json.Unmarshal(bytes, &respBody)
	if err != nil {
		*hrl.appResponseCode = respBody.ResponseCode
	}

	hrl.logger.Info(hrl.context, "Http Response",
		zap.String(ins.LabelHTTPService, path),
		zap.Any(ins.LabelHTTPResponse, makemapresp),
		zap.Any(ins.LabelHTTPStatus, hrl.writer.Status()),
		zap.Any(ins.LabelAppResponseCode, respBody.ResponseCode),
	)

	return n, nil
}

func (hrl *httpResponseLogger) WriteHeader(statusCode int) {
	hrl.writer.WriteHeader(statusCode)
	hrl.logger.Info(hrl.context, "Http status", zap.Int(ins.LabelHTTPStatus, statusCode))
}
