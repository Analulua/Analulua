package grpc

import (
	"context"
	"log"
	"strings"

	cctx "newdemo1/resource/jaeger/common/context"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

const (
	bearer        string = "bearer"
	authorization string = "authorization"
)

func extractTokenFromAuthHeader(val string) (token string, ok bool) {
	authHeaderParts := strings.Split(val, " ")
	if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], bearer) {
		return "", false
	}

	return authHeaderParts[1], true
}

func extractUserInfo(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	authHeader, ok := md[authorization]
	if !ok {
		return ctx
	}

	token, ok := extractTokenFromAuthHeader(authHeader[0])
	if ok {
		parser := jwt.Parser{}
		claims := jwt.MapClaims{}
		_, _, err := parser.ParseUnverified(token, claims)
		if err != nil {
			log.Println(err)
			return ctx
		}
		mobile, ok := claims[cctx.CtxMobile.String()]
		if ok {
			ctx = context.WithValue(ctx, cctx.CtxMobile, mobile)
		}
		userID, ok := claims[cctx.CtxUserID.String()]
		if ok {
			ctx = context.WithValue(ctx, cctx.CtxUserID, userID)
		}
	}

	return ctx
}
