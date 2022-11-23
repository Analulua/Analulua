package instrumentation

import "go.opentelemetry.io/otel/label"

const (
	XRequestID    = "x-request-id"
	XForwardedFor = "x-forwarded-for"
	UserAgent     = "user-agent"
	ErrorCode     = "error_code"
	ErrorMessage  = "error_message"
)

const (
	Method          = label.Key("method")
	CorrelationID   = label.Key("correlation.id")
	ResponseCode    = label.Key("response.bussiness.code")
	ResponseMessage = label.Key("response.bussiness.message")
)

const (
	LabelStatus           = "response.status"
	LabelMethod           = "request"
	LabelHTTPHeader       = "http.header"
	LabelHTTPRequest      = "http.request"
	LabelHTTPResponse     = "http.response"
	LabelHTTPStatus       = "http.status"
	LabelCorrelationID    = "correlation.id"
	LabelGRPCHeader       = "rpc.header"
	LabelGRPCRequest      = "rpc.request"
	LabelGRPCResponse     = "rpc.response"
	LabelGRPCService      = "rpc.service"
	LabelGRPCResponseCode = "rpc.response_code"
	LabelHTTPService      = "http.service"
	LabelClientIP         = "client.ip"
	LabelUserAgent        = "http.user_agent"
	LabelHTTPMethod       = "http.method"
	LabelAppResponseCode  = "app.response_code"
)

const (
	ElapseNumber = 1000
)
