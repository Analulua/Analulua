package constant

import (
	"google.golang.org/grpc/codes"
	"net/http"
	commonErr "newdemo1/resource/jaeger/common/error"
)

var (
	Success                          = commonErr.ServiceError{Code: "000", Message: "Success"}
	ServiceErrorCodeToHttpStatusCode = map[string]int{
		Success.Code: http.StatusOK,
	}

	ServiceErrorCodeToGRPCErrorCode = map[string]codes.Code{
		Success.Code: codes.OK,
	}
)
