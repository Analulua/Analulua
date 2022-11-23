package instrumentation

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"google.golang.org/grpc/status"
	commonError "newdemo1/resource/jaeger/common/error"
	commongrpc "newdemo1/resource/jaeger/common/grpc"
)

const (
	MetricResponseCodeSuccess = "response_code:0_success"
	MetricResponseCode        = "response_code"
)

var metricNames = make(map[string]string)

/**
AddMetricNameByServiceMethod add metrics by service method
in => service method
v => metrics name

Eg:-
k = /proto.CustomerHandler/RegisterCustomer
v: customer.register
*/
func AddMetricNameByGRPCServiceMethod(in interface{}, v string) {
	fullM := getFullMethodNameFromServerHandler(in)
	metricNames[fullM] = v
}

/**
AddMetricName
k => service method name
v => metrics name
*/
func AddMetricName(k, v string) {
	metricNames[k] = v
}

func GetResponseCode(err error) string {
	sErr := ErrorMapping(err)
	m := fmt.Sprintf("%s:%s", MetricResponseCode, sErr.Code)
	return m
}

// input => git.capitalx.id/dimii/customer/proto.CustomerHandlerServer.RegisterCustomer
// output => /proto.CustomerHandler/RegisterCustomer
func getFullMethodNameFromServerHandler(m interface{}) string {
	s1 := runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
	i := strings.LastIndex(s1, "/")
	s2 := s1[i:]
	return strings.Replace(s2, "Server.", "/", 1)
}

func ErrorMapping(err error) commonError.ServiceError {
	status, _ := status.FromError(err)
	d := status.Details()
	if len(d) > 0 {
		details := d[0].(*commongrpc.Error)
		return commonError.ServiceError{
			Code:       details.Code,
			Message:    status.Message(),
			Attributes: details.Attributes,
		}
	}
	return commonError.ServiceError{Message: status.Message()}
}
