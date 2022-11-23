package grpc

import (
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/semconv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	svcerr "newdemo1/resource/jaeger/common/error"
	"newdemo1/resource/jaeger/common/telemetry"
	ins "newdemo1/resource/jaeger/common/telemetry/instrumentation"
	"newdemo1/resource/jaeger/common/telemetry/instrumentation/device"
	"newdemo1/resource/jaeger/common/telemetry/instrumentation/filter"
)

var propagator = otel.NewCompositeTextMapPropagator(propagators.TraceContext{},
	b3.B3{InjectEncoding: b3.B3MultipleHeader | b3.B3SingleHeader})
var options = []otelgrpc.Option{otelgrpc.WithPropagators(propagator)}

type messageType label.KeyValue

// Event adds an event of the messageType to the span associated with the
// passed context with id and size (if message is a proto message).
func (m messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent(ctx, "message",
			label.KeyValue(m),
			semconv.RPCMessageIDKey.Int(id),
			semconv.RPCMessageUncompressedSizeKey.Int(proto.Size(p)),
		)
	} else {
		span.AddEvent(ctx, "message",
			label.KeyValue(m),
			semconv.RPCMessageIDKey.Int(id),
		)
	}
}

var (
	messageSent     = messageType(semconv.RPCMessageTypeSent)
	messageReceived = messageType(semconv.RPCMessageTypeReceived)
)

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor suitable
// for use in a grpc.Dial call.
func UnaryClientInterceptor(api *telemetry.API, errorMapper map[string]grpcCodes.Code) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		// TODO: ask why incoming metadata need to be propagated
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		outgoingMetadata, _ := metadata.FromOutgoingContext(ctx)

		metadataCopy := requestMetadata.Copy()
		metadataCopy = metadata.Join(metadataCopy, outgoingMetadata)

		tracer := api.Tracer(trace.WithInstrumentationVersion(otelcontrib.SemVersion()))

		name, attr := spanInfo(method, cc.Target())
		var span trace.Span
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		otelgrpc.Inject(ctx, &metadataCopy, options...)
		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		startTime := time.Now()

		messageSent.Event(ctx, 1, req)

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		messageReceived.Event(ctx, 1, reply)

		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		}

		isInboundTraffic := false
		pushAdditional(ctx, api, name, method, err, req, reply, startTime, isInboundTraffic, errorMapper)
		return err
	}
}

type streamEventType int

type streamEvent struct {
	Type streamEventType
	Err  error
}

const (
	closeEvent streamEventType = iota
	receiveEndEvent
	errorEvent
)

// clientStream  wraps around the embedded grpc.ClientStream, and intercepts the RecvMsg and
// SendMsg method call.
type clientStream struct {
	grpc.ClientStream

	desc       *grpc.StreamDesc
	events     chan streamEvent
	eventsDone chan struct{}
	finished   chan error

	receivedMessageID int
	sentMessageID     int
}

var _ = proto.Marshal

func (w *clientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	switch {
	case err == nil && !w.desc.ServerStreams:
		w.sendStreamEvent(receiveEndEvent, nil)
	case err == io.EOF:
		w.sendStreamEvent(receiveEndEvent, nil)
	case err != nil:
		w.sendStreamEvent(errorEvent, err)
	default:
		w.receivedMessageID++
		messageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *clientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)

	w.sentMessageID++
	messageSent.Event(w.Context(), w.sentMessageID, m)

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (w *clientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return md, err
}

func (w *clientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	} else {
		w.sendStreamEvent(closeEvent, nil)
	}

	return err
}

const (
	clientClosedState byte = 1 << iota
	receiveEndedState
)

func wrapClientStream(s grpc.ClientStream, desc *grpc.StreamDesc) *clientStream {
	events := make(chan streamEvent)
	eventsDone := make(chan struct{})
	finished := make(chan error)

	go func() {
		defer close(eventsDone)

		// Both streams have to be closed
		state := byte(0)

		for event := range events {
			switch event.Type {
			case closeEvent:
				state |= clientClosedState
			case receiveEndEvent:
				state |= receiveEndedState
			case errorEvent:
				finished <- event.Err
				return
			}

			if state == clientClosedState|receiveEndedState {
				finished <- nil
				return
			}
		}
	}()

	return &clientStream{
		ClientStream: s,
		desc:         desc,
		events:       events,
		eventsDone:   eventsDone,
		finished:     finished,
	}
}

func (w *clientStream) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-w.eventsDone:
	case w.events <- streamEvent{Type: eventType, Err: err}:
	}
}

// StreamClientInterceptor returns a grpc.StreamClientInterceptor suitable
// for use in a grpc.Dial call.
func StreamClientInterceptor(api *telemetry.API, errorMapper map[string]grpcCodes.Code) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// TODO: ask why incoming metadata need to be propagated
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		outgoingMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		metadataCopy = metadata.Join(metadataCopy, outgoingMetadata)

		tracer := api.Tracer(trace.WithInstrumentationVersion(otelcontrib.SemVersion()))

		name, attr := spanInfo(method, cc.Target())
		var span trace.Span
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attr...),
		)

		otelgrpc.Inject(ctx, &metadataCopy, options...)
		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		startTime := time.Now()

		s, err := streamer(ctx, desc, cc, method, callOpts...)
		stream := wrapClientStream(s, desc)

		go func() {
			if err == nil {
				err = <-stream.finished
			}

			if err != nil {
				s, _ := status.FromError(err)
				span.SetStatus(codes.Error, s.Message())
			}

			span.End()
		}()

		isInboundTraffic := false
		pushAdditional(ctx, api, name, method, err, nil, nil, startTime, isInboundTraffic, errorMapper)
		return stream, err
	}
}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor suitable
// for use in a grpc.NewServer call.
func UnaryServerInterceptor(api *telemetry.API, errorMapper map[string]grpcCodes.Code) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		entries, spanCtx := otelgrpc.Extract(ctx, &metadataCopy, options...)
		ctx = otel.ContextWithBaggageValues(ctx, entries...)

		tracer := api.Tracer(trace.WithInstrumentationVersion(otelcontrib.SemVersion()))

		name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
		ctx, span := tracer.Start(
			trace.ContextWithRemoteSpanContext(ctx, spanCtx),
			name,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		startTime := time.Now()

		messageReceived.Event(ctx, 1, req)

		resp, err := handler(ctx, req)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			messageSent.Event(ctx, 1, s.Proto())
		} else {
			messageSent.Event(ctx, 1, resp)
		}

		isInboundTraffic := true
		pushAdditional(ctx, api, name, info.FullMethod, err, req, resp, startTime, isInboundTraffic, errorMapper)
		return resp, err
	}
}

// serverStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type serverStream struct {
	grpc.ServerStream
	ctx context.Context

	receivedMessageID int
	sentMessageID     int
}

func (w *serverStream) Context() context.Context {
	return w.ctx
}

func (w *serverStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	if err == nil {
		w.receivedMessageID++
		messageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *serverStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)

	w.sentMessageID++
	messageSent.Event(w.Context(), w.sentMessageID, m)

	return err
}

func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

// StreamServerInterceptor returns a grpc.StreamServerInterceptor suitable
// for use in a grpc.NewServer call.
func StreamServerInterceptor(api *telemetry.API, errorMapper map[string]grpcCodes.Code) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		entries, spanCtx := otelgrpc.Extract(ctx, &metadataCopy, options...)
		ctx = otel.ContextWithBaggageValues(ctx, entries...)

		tracer := api.Tracer(trace.WithInstrumentationVersion(otelcontrib.SemVersion()))

		name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
		ctx, span := tracer.Start(
			trace.ContextWithRemoteSpanContext(ctx, spanCtx),
			name,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		startTime := time.Now()

		err := handler(srv, wrapServerStream(ctx, ss))
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		}
		isInboundTraffic := true
		pushAdditional(ctx, api, name, info.FullMethod, err, nil, nil, startTime, isInboundTraffic, errorMapper)
		return err
	}
}

// spanInfo returns a span name and all appropriate attributes from the gRPC
// method and peer address.
func spanInfo(fullMethod, peerAddress string) (string, []label.KeyValue) {
	attrs := []label.KeyValue{semconv.RPCSystemGRPC}
	name, mAttrs := parseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(peerAddress)...)
	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []label.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []label.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []label.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

// peerFromCtx returns a peer address from a context, if one exists.
func peerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

// parseFullMethod returns a span name following the OpenTelemetry semantic
// conventions as well as all applicable span label.KeyValue attributes based
// on a gRPC's FullMethod.
func parseFullMethod(fullMethod string) (string, []label.KeyValue) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	lenParts := 2
	if len(parts) != lenParts {
		// Invalid format, does not follow `/package.service/method`.
		return name, []label.KeyValue(nil)
	}

	var attrs []label.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}
	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}
	return name, attrs
}

func pushAdditional(ctx context.Context,
	api *telemetry.API,
	name, method string, err error,
	req interface{}, resp interface{},
	start time.Time,
	isInbound bool,
	errorMapper map[string]grpcCodes.Code) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	stat := uint32(status.Code(err))
	if api.ServiceAPI != "" {
		name = api.ServiceAPI
	}

	src_env := api.SourceEnv

	var responseCode string = "0"
	var appResponseCode string = "N/A"
	elapsedTime := time.Since(start).Milliseconds()

	if err != nil {
		switch serviceError := err.(type) {
		case svcerr.ServiceError:

			appResponseCode = serviceError.Code
			grpcCode := errorMapper[serviceError.Code]
			responseCode = strconv.FormatInt(int64(grpcCode), 10)

		default:
			responseCode = "2" // Unknown
		}
	}

	// send metrics to datadog
	rsCode := []string{
		"grpc_endpoint:" + method,
		"src_env:" + src_env,
		"response_code:" + responseCode,
		"error_message:" + errorMessage,
		"app_response_code:" + appResponseCode,
	}

	metricsName := api.ServiceAPI
	if !isInbound {
		metricsName = metricsName + ".external_grpc"
	}

	api.Metric().Count(metricsName, 1, rsCode)
	api.Metric().Histogram(metricsName+".histogram", float64(elapsedTime), rsCode)
	api.Metric().Distribution(metricsName+".distribution", float64(elapsedTime), rsCode)

	// send metrics to open telemetry
	vr, err := api.Meter().NewInt64ValueRecorder(name + ".valuerecorder")
	if err == nil {
		vr.Record(ctx, elapsedTime, label.Uint32(ins.LabelStatus, stat), label.String(ins.LabelMethod, method))
	}

	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()

	span := trace.SpanFromContext(ctx)

	// add to trace attributes
	attributes := device.Extract(&metadataSupplier{metadata: &metadataCopy})
	attributes = append(attributes, label.Uint32(ins.LabelStatus, stat))
	attributes = append(attributes,
		label.String(ins.LabelCorrelationID,
			span.SpanContext().TraceID.String()+"-"+span.SpanContext().SpanID.String()),
	)
	span.SetAttributes(attributes...)

	// extract grpc metadata to trace attributes
	extractHeader(&metadataCopy, ins.XForwardedFor, semconv.HTTPClientIPKey, span)
	extractHeader(&metadataCopy, ins.UserAgent, semconv.HTTPUserAgentKey, span)
	extractHeader(&metadataCopy, ins.ErrorCode, ins.ResponseCode, span)
	extractHeader(&metadataCopy, ins.ErrorMessage, ins.ResponseMessage, span)

	metadataFilters := filter.DefaultHeaderFilter

	filteredRequest := req
	filteredResponse := resp
	if api.Filter != nil {
		rules := api.Filter.PayloadFilter(&filter.TargetFilter{
			Method: method,
		})

		filteredRequest = filter.BodyFilter(rules, req)
		filteredResponse = filter.BodyFilter(rules, resp)
		metadataFilters = append(metadataFilters, api.Filter.HeaderFilter...)
	}

	strMetadata := filter.MetadataFilter(metadataCopy, metadataFilters)
	// add log for request,response,header
	api.Logger().Info(ctx, "",
		zap.Any(ins.LabelGRPCHeader, strMetadata),
		zap.Any(ins.LabelGRPCRequest, filteredRequest),
		zap.Any(ins.LabelGRPCResponse, filteredResponse),
		zap.String(ins.LabelGRPCService, method),
		zap.String(ins.LabelGRPCResponseCode, responseCode),
		zap.String(ins.LabelAppResponseCode, appResponseCode),
	)
}

type metadataSupplier struct {
	metadata *metadata.MD
}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (s *metadataSupplier) Set(key, value string) {
	s.metadata.Set(key, value)
}

func extractHeader(metadata *metadata.MD, header string, key label.Key, span trace.Span) {
	val := metadata.Get(header)
	if len(val) > 0 {
		span.SetAttributes(key.String(val[0]))
	}
}

func extractMetadata(metadata *metadata.MD, header string) string {
	val := metadata.Get(header)
	if len(val) > 0 {
		return val[0]
	}
	return ""
}
