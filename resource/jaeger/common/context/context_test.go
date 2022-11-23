package context

import (
	"context"
	"testing"
)

func TestContextKey_String(t *testing.T) {
	if got := CtxMobile.String(); got == "" {
		t.Fatalf("bad context: %v", got)
	}
}

func TestGetContextAsString(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxUserID, 1)
	ctx = context.WithValue(ctx, CtxMobile, "081234567890")

	if got := GetContextAsString(ctx, CtxMobile); got == "" {
		t.Fatalf("bad context: %v", got)
	}
}
