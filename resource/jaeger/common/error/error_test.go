package error

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError(t *testing.T) {
	svcError := ServiceError{
		Code:       "INVALID_LOGIN",
		Message:    "Invalid user or password",
		Attributes: nil,
	}

	if got := svcError.Error(); got == "" {
		t.Fatalf("bad error: %v", got)
	}
}

func TestServiceError_Comparable(t *testing.T) {
	svcError := ServiceError{
		Code:       "INVALID_LOGIN",
		Message:    "Invalid user or password",
		Attributes: nil,
	}

	assert.ErrorIs(t, svcError, svcError)
}
