package grpc

import (
	"context"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestExtractTokenFromAuthHeader(t *testing.T) {
	if _, ok := extractTokenFromAuthHeader("Bearer token"); !ok {
		t.Fatal("bad auth header format")
	}
}

func TestExtractUserInfo(t *testing.T) {
	md := make(map[string][]string)
	md[authorization] = []string{"Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6InB1YmxpYzpmMzQxNDIzMi1hMDYwLTRiY2UtYjc4Ni1lYjU4MWVjOGViMWIiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2MzQ0MjcwNDUsImlhdCI6MTYwMjg5MTA0NSwiaXNzIjoiYXBpLmRpbWlpLmRldiIsIm1vYmlsZSI6IjA4MTM3Njc2Mzc4NCIsIm5hbWUiOiJUZXN0Iiwic3ViIjoidGVzdCIsInVzZXJfaWQiOiJ0ZXN0In0.m5n_CVlTkrzCVaaVGh-ngz97qLi7i4armKS-SyDJeMECMyfp_5wUQKhqUrkcv4m1AziqwTqBES1L_mbwZem5zyAaUgHPtPW7hqZztBGt1GBc1N-ReYgCU90VUzb0TZ7aWfdn3C1IiN8IAfT9ODtxYrSego5B676wjszbHaR-6wU7_IE8UOIHmKJ8F0oSp3UZEtuSBhMph3jwe5-FSis_FqWh8XE5ED8Bf4ITAeFdX7hCjtdNjN5xoSIEFEFD5apUhyyzCXJQvzDGXyFesC-HnHmwXEnsKLDu9bPlMOvREnOQ64LmlxsAyce6_Rp5pSnRTKNdqqfSRyokQItSFD1dSJ8oGTjOzaNqgkoJI70oSwQL7DDS_LZADzXSOyTtILKVShoezx3SZ6XUt984gfq7m1L7tP7SLWacWe5Bn4KY5GxNITw6uYGDhRNG_JNcswV3lDDZq1gXm5Ck8Ys9RaLHDqXhoIEIAIGGeQJp6zhh_yQx8p6R-4BEbKoojB9H0-rHj8ijE00FGlUYAVaZs3HeFpuebV2aWUuhBqCV5LN-kwIw6KQo0mKU7o2MMCCmABRPjJ1DS7y-2rjHA_LXuJz0lz9RZ-Nd1FpcNb8AyOdRGK7yzP2gCIQ-5ffpS4_xoEcLG00db843PlkS8vJecnaiNXQXgoYQvoZatYAv64JHnK4"}
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, md)

	if got := extractUserInfo(ctx); got == ctx {
		t.Fatalf("bad context: %v ", got)
	}
}
