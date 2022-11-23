package device

import (
	"go.opentelemetry.io/otel/label"
)

const (
	deviceID   = "deviceId"
	lat        = "latitude"
	long       = "longitude"
	appVersion = "appVersion"
	osName     = "osName"
	osVersion  = "osVersion"
	deviceName = "deviceName"
	lang       = "lang"
	isRooted   = "isRooted"
)

var info = []string{deviceID, lat, long, appVersion, osName, osVersion, deviceName, lang, isRooted}

// MapCarrier is the storage medium used by a TextMapPropagator.
type MapCarrier interface {
	// Get returns the value associated with the passed key.
	Get(key string) string
	// Set stores the key-value pair.
	Set(key string, value string)
}

func Extract(carrier MapCarrier) []label.KeyValue {
	var d []label.KeyValue

	for _, f := range info {
		val := carrier.Get(f)
		if len(val) != 0 {
			d = append(d, label.String("device."+f, val))
		}
	}

	return d
}
