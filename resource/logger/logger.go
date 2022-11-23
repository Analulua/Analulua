package logger

import (
	"newdemo1/resource/jaeger/common/telemetry"
)

type (
	Logger struct {
		telemetry.Logger
	}
)

func NewLogger(api *telemetry.API) (Logger, error) {
	return Logger{
		Logger: api.Logger(),
	}, nil
}
