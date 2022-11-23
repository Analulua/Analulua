package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	api "newdemo1/resource/jaeger/common/telemetry"
)

type metrics struct {
	trip   http.RoundTripper
	metric MetricType
	api    *api.API
	ctx    context.Context
	xml    bool // default is false (json)
	resp   interface{}
}

func WithMetrics(ctx context.Context, trip http.RoundTripper, metricType MetricType, api *api.API,
	xml bool, resp interface{}) http.RoundTripper {

	return &metrics{
		trip:   trip,
		metric: metricType,
		api:    api,
		ctx:    ctx,
		xml:    xml,
		resp:   resp,
	}
}

func (m *metrics) RoundTrip(req *http.Request) (*http.Response, error) {

	start := time.Now()

	res, err := m.trip.RoundTrip(req)

	elapsedTime := time.Since(start).Milliseconds()

	if err != nil {
		sendMetric("error", m.metric, elapsedTime, m.api)
	} else {
		if m.metric.StatusPath != "" && res.Body != nil {
			body, errs := ioutil.ReadAll(res.Body)
			if errs != nil {
				m.api.Logger().Error(m.ctx, "[common:http:httpClient] ioutil.ReadAll Error:", errs)
				sendMetric(strconv.Itoa(res.StatusCode), m.metric, elapsedTime, m.api)
				return res, errs
			}
			// rewrite the body since ReadAll close the Reader
			res.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			m.api.Logger().Info(m.ctx, fmt.Sprintf("[common:http:httpClient] HTTP request Resp:%s", string(body)))

			if m.xml {
				if errXML := xml.Unmarshal(body, m.resp); errXML != nil {
					m.api.Logger().Error(m.ctx, "[common:http:httpClient] XML unmarshal Error:", errXML)
					sendMetric(strconv.Itoa(res.StatusCode), m.metric, elapsedTime, m.api)
					return res, errXML
				}
			} else {
				if errJSON := json.Unmarshal(body, m.resp); errJSON != nil {
					m.api.Logger().Error(m.ctx, "[common:http:httpClient] JSON unmarshal Error:", errJSON)
					sendMetric(strconv.Itoa(res.StatusCode), m.metric, elapsedTime, m.api)
					return res, errJSON
				}
			}

			sc := getHTTPRespStatus(m.ctx, m.metric.StatusPath, m.resp, m.api.Logger())
			// log metrics
			if sc == "" {
				// if no application status code then http status
				sc = strconv.Itoa(res.StatusCode)
			}
			sendMetric(sc, m.metric, elapsedTime, m.api)
		} else {

			sendMetric(strconv.Itoa(res.StatusCode), m.metric, elapsedTime, m.api)
		}
	}

	return res, err
}

func sendMetric(sc string, metric MetricType, latency int64, api *api.API) {
	status := fmt.Sprintf("status:%v", sc)
	tags := append(metric.Tags, status)
	api.Metric().Incr(metric.MetricName, tags)
	api.Metric().Histogram(metric.MetricName+".latency", float64(latency), tags)
}

type MetricType struct {
	MetricName string
	Tags       []string
	StatusPath string
}

/**
GetHTTPRespStatus Helper functon to extract a field value as specifed in the path

Eg:-  path =  "Result.[0].Status" means we will extract value from that struct path
*/
func getHTTPRespStatus(ctx context.Context, path string, resp interface{}, logger api.Logger) string {
	fields := strings.Split(path, ".")
	v := reflect.ValueOf(resp).Elem()

	for i := 0; i < len(fields) && v.IsValid(); i++ {
		v = handleReflection(ctx, fields[i], v, logger)
	}

	var r string
	if v.IsValid() {
		r = fmt.Sprintf("%v", v.Interface())
	}
	return r
}

func handleReflection(ctx context.Context, fieldName string, v reflect.Value, logger api.Logger) reflect.Value {
	switch v.Kind() {
	case reflect.Invalid:
		logger.Info(ctx, "[common:metrics:handleReflection] Invalid reflection")
	case reflect.Slice, reflect.Array:
		if v.Len() > 0 {
			v = v.Index(0)
		}
		return v
	case reflect.Struct:
		if v.NumField() > 0 {
			v = v.FieldByName(fieldName)
		}
		return v
	default: // basic types, channels, funcs
		return v
	}
	return v
}

func MetricsResponseMiddleware(ctx context.Context, metricType MetricType, api *api.API,
	isXML bool, resp interface{}) resty.ResponseMiddleware {

	return func(c *resty.Client, res *resty.Response) error {
		body := res.Body()
		elapsedTime := res.Time().Milliseconds()
		if metricType.StatusPath != "" && len(body) > 0 {

			api.Logger().Info(ctx, fmt.Sprintf("[common:http:httpClient] HTTP request Resp:%s", string(body)))

			if isXML {
				if errXML := xml.Unmarshal(body, resp); errXML != nil {
					api.Logger().Error(ctx, "[common:http:httpClient] XML unmarshal Error:", errXML)
					sendMetric(strconv.Itoa(res.StatusCode()), metricType, elapsedTime, api)
					return errXML
				}
			} else {
				if errJSON := json.Unmarshal(body, resp); errJSON != nil {
					api.Logger().Error(ctx, "[common:http:httpClient] JSON unmarshal Error:", errJSON)
					sendMetric(strconv.Itoa(res.StatusCode()), metricType, elapsedTime, api)
					return errJSON
				}
			}

			sc := getHTTPRespStatus(ctx, metricType.StatusPath, resp, api.Logger())
			// log metrics
			if sc == "" {
				// if no application status code then http status
				sc = strconv.Itoa(res.StatusCode())
			}
			sendMetric(sc, metricType, elapsedTime, api)
		} else {

			sendMetric(strconv.Itoa(res.StatusCode()), metricType, elapsedTime, api)
		}

		return nil
	}
}
