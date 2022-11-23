package context

import (
	gctx "context"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	// CtxUserID is context key for user id
	CtxUserID = contextKey("user_id")
	// CtxMobile is context key for mobile number
	CtxMobile = contextKey("mobile")
)

// GetContextString return context value as type string
func GetContextAsString(ctx gctx.Context, ctxKey contextKey) string {
	val := ctx.Value(ctxKey)
	if val != nil {
		str, ok := val.(string)
		if ok {
			return str
		}
	}
	return ""
}
