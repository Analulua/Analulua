package filter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/grpc/metadata"
)

var (
	// DefaultHeaderFilter default value for filter header
	DefaultHeaderFilter = []string{"authorization", "Authorization"}
)

type TargetFilter struct {
	Method string
}

func (c *TargetFilter) GetMethod() string {
	return c.Method
}

func BodyFilter(re []*regexp.Regexp, paylod interface{}) interface{} {
	item := fmt.Sprintf("%v", paylod)
	if item == "" {
		return ""
	}

	for _, regex := range re {
		value := regex.FindStringSubmatch(item)
		maxLength := len(value)
		if len(value) > 0 {
			item = strings.ReplaceAll(item, value[maxLength-1], "***")
		}
	}

	return item
}

func MetadataFilter(item metadata.MD, filters []string) string {
	if len(item) == 0 {
		return ""
	}

	jsonHeader := make(map[string]interface{})
	for headerKey, val := range item {
		strValue := strings.Join(val, ",")
		jsonHeader[headerKey] = strValue
	}

	return processFilter(jsonHeader, filters)
}

func HeaderFilter(payload http.Header, filters []string) string {
	if len(payload) == 0 {
		return ""
	}

	jsonHeader := make(map[string]interface{})
	for headerKey, val := range payload {
		strValue := strings.Join(val, ",")
		jsonHeader[headerKey] = strValue
	}

	return processFilter(jsonHeader, filters)
}

func processFilter(item map[string]interface{}, filters []string) string {
	for _, filter := range filters {
		_, ok := item[filter]
		if ok {
			item[filter] = "***"
		}
	}

	data, err := json.Marshal(item)
	if err != nil {
		return ""
	}

	return string(data)
}
