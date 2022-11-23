package tracer

import (
	"context"
	"encoding/json"
	"strconv"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
)

type Tracer interface {
	Context() context.Context
	Finish(additionalTags ...map[string]interface{})
}

type tracerImpl struct {
	ctx  context.Context
	span trace.Span
	tags map[string]interface{}
}

func CloneTrace(fromCtx context.Context, toCtx context.Context) context.Context {
	span := trace.SpanFromContext(fromCtx)
	spanContext := span.SpanContext()
	if spanContext.HasTraceID() {
		toCtx = trace.ContextWithSpan(toCtx, span)
	}

	return toCtx
}

func StartTrace(ctx context.Context, opsName string) Tracer {
	tr := global.Tracer(opsName)
	ctx = trace.ContextWithRemoteSpanContext(ctx, trace.RemoteSpanContextFromContext(ctx))

	var span trace.Span
	ctx, span = tr.Start(ctx, opsName)

	return &tracerImpl{
		span: span,
		ctx:  ctx,
	}
}

func (t *tracerImpl) SetError(err error) {
	if err == nil {
		return
	}
	t.span.SetStatus(1, err.Error())
	t.span.RecordError(t.Context(), err)
}

func (t *tracerImpl) Context() context.Context {
	return t.ctx
}

func (t *tracerImpl) Finish(additionalTags ...map[string]interface{}) {
	defer t.span.End()

	if additionalTags != nil && t.tags == nil {
		t.tags = make(map[string]interface{})
	}

	for _, tag := range additionalTags {
		for k, v := range tag {
			t.tags[k] = v
		}
	}

	for k, v := range t.tags {
		t.span.SetAttributes(label.Key(k).String(toString(v)))
	}
}

func toString(v interface{}) (s string) {
	switch val := v.(type) {
	case error:
		if val != nil {
			s = val.Error()
		}
	case string:
		s = val
	case int:
		s = strconv.Itoa(val)
	default:
		b, _ := json.Marshal(val)
		s = string(b)
	}

	return
}
