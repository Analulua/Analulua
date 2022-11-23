//go:generate protoc -I . --go_out=. --go_opt=module=newdemo1/resource/jaeger/common/grpc aladinbank_internal_error.proto
package grpc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"newdemo1/resource/jaeger/common/crypto"
	svcerr "newdemo1/resource/jaeger/common/error"
	"newdemo1/resource/jaeger/common/tls"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
)

const (
	errorCode    = "error_code"
	errorMessage = "error_message"
	signature    = "signature"
	keyID        = "key_id"
)

func errorMetadata(serviceError svcerr.ServiceError) metadata.MD {
	data := make(map[string]string)
	data[errorCode] = serviceError.Code
	data[errorMessage] = serviceError.Message
	if len(serviceError.Attributes) > 0 {
		for k, v := range serviceError.Attributes {
			data[k] = v
		}
	}
	return metadata.New(data)
}

func grpcError(statusCode codes.Code, serviceError svcerr.ServiceError) error {
	pbError := Error{
		Code:       serviceError.Code,
		Message:    serviceError.Message,
		Attributes: serviceError.Attributes,
	}
	grpcError := status.New(statusCode, serviceError.Message)
	errWithDetails, err := grpcError.WithDetails(&pbError)
	if err != nil {
		return grpcError.Err()
	}
	return errWithDetails.Err()
}

// UnaryErrorInterceptor returns a new unary server interceptor that added error detail for service error.
func UnaryErrorInterceptor(errorMapper map[string]codes.Code) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			switch serviceError := err.(type) {
			case svcerr.ServiceError:
				md := errorMetadata(serviceError)
				if err := grpc.SetTrailer(ctx, md); err != nil {
					log.Print(err)
				}

				statusCode, found := errorMapper[serviceError.Code]

				if !found {
					statusCode = codes.Unknown
				}

				return nil, grpcError(statusCode, serviceError)
			default:
				return nil, err
			}
		}

		return resp, nil
	}
}

// StreamErrorInterceptor returns a new streaming server interceptor that added error detail for service error.
func StreamErrorInterceptor(errorMapper map[string]codes.Code) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			switch serviceError := err.(type) {
			case svcerr.ServiceError:
				md := errorMetadata(serviceError)
				stream.SetTrailer(md)

				statusCode, found := errorMapper[serviceError.Code]

				if !found {
					statusCode = codes.Unknown
				}

				return grpcError(statusCode, serviceError)
			default:
				return err
			}
		}

		return nil
	}
}

// UnaryAuthInterceptor returns a new unary server interceptor that extract user info from token.
func UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		ctx = extractUserInfo(ctx)
		return handler(ctx, req)
	}
}

func recoveryInterceptor() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	handler := func(p interface{}) (err error) {
		log.Print("panic", p)
		return status.Error(codes.Internal, "server panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(handler),
	}

	return recovery.UnaryServerInterceptor(opts...), recovery.StreamServerInterceptor(opts...)
}

// WithRecovery return gRPC server options with recovery handler
func WithRecovery() []grpc.ServerOption {
	unary, stream := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(unary),
		grpc.StreamInterceptor(stream),
	}
	return serverOptions
}

// WithValidation returns gRPC server options with request validator
func WithValidation() []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(validator.UnaryServerInterceptor()),
		grpc.StreamInterceptor(validator.StreamServerInterceptor()),
	}
	return serverOptions
}

// WithErrorDetails returns gRPC server options with request validator
func WithErrorDetails(errorMapper map[string]codes.Code) []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(UnaryErrorInterceptor(errorMapper)),
		grpc.StreamInterceptor(StreamErrorInterceptor(errorMapper)),
	}
	return serverOptions
}

// WithSecure returns gRPC server option with SSL credentials
func WithSecure(ca, cert, key []byte, mutual bool) grpc.ServerOption {
	tlsCfg := tls.WithCertificate(ca, cert, key, mutual)
	return grpc.Creds(credentials.NewTLS(tlsCfg))
}

// DefaultServerOptions returns default gRPC server option with validation and recovery
func WithDefault(errorMapper map[string]codes.Code) []grpc.ServerOption {
	unaryRecovery, streamRecovery := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(),
			UnaryAuthInterceptor(),
			unaryRecovery,
			UnaryErrorInterceptor(errorMapper),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(),
			streamRecovery,
			StreamErrorInterceptor(errorMapper),
		)}
	return serverOptions
}

type KeyStore interface {
	GetSecretKey(string) []byte
}

type SignatureInfo struct {
	Alg             string
	HeaderFieldKeys string
	BodyFieldKeys   string
	DefaultFields   []string
}

func getValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	value, ok := md[key]
	if !ok {
		return ""
	}

	if len(value) > 0 {
		return value[0]
	}

	return ""
}

func getClaimValue(ctx context.Context, key string) string {
	auth := getValue(ctx, authorization)
	token, ok := extractTokenFromAuthHeader(auth)
	if ok {
		parser := jwt.Parser{}
		claims := jwt.MapClaims{}
		_, _, err := parser.ParseUnverified(token, claims)
		if err != nil {
			log.Println(err)
			return ""
		}
		value, ok := claims[key]
		if ok {
			return value.(string)
		}
	}
	return ""
}

func constructSignatureFields(ctx context.Context, req interface{}, headerFields, bodyFields string,
	defaultFields []string) string {

	var fields strings.Builder

	if len(headerFields) > 0 {
		headers := strings.Split(headerFields, " ")
		for _, header := range headers {
			fields.WriteString(getValue(ctx, header))
		}
	}

	if len(bodyFields) > 0 {
		bodies := strings.Split(bodyFields, " ")
		for _, body := range bodies {
			request := req.(proto.Message)
			fieldDesc := request.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(body))
			if fieldDesc != nil {
				value := request.ProtoReflect().Get(fieldDesc)
				if value.IsValid() {
					fields.WriteString(value.String())
				}
			}
		}
	}

	if fields.Len() == 0 {
		if len(defaultFields) > 0 {
			for _, header := range defaultFields {
				fields.WriteString(getValue(ctx, header))
			}
		} else {
			msg := req.(fmt.Stringer)
			fields.WriteString(msg.String())
		}
	}

	return fields.String()
}

func UnarySignatureInterceptor(keyStore KeyStore, signInfo SignatureInfo) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		headerFields := getValue(ctx, signInfo.HeaderFieldKeys)
		bodyFields := getValue(ctx, signInfo.BodyFieldKeys)

		fields := constructSignatureFields(ctx, req, headerFields, bodyFields, signInfo.DefaultFields)

		keyID := getClaimValue(ctx, keyID)
		secretKey := keyStore.GetSecretKey(keyID)

		var hash string
		switch signInfo.Alg {
		case "HS256":
			hash = crypto.HMAC(sha256.New, secretKey, fields)
		default:
			hash = crypto.HMAC(sha256.New, secretKey, fields)
		}

		requestHash := getValue(ctx, signature)
		if hash != requestHash {
			return nil, status.Error(codes.InvalidArgument, "invalid signature")
		}

		resp, err := handler(ctx, req)
		return resp, err
	}
}

// WithSignature returns gRPC server options with signature validator
func WithSignature(keyStore KeyStore, signInfo SignatureInfo) []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(UnarySignatureInterceptor(keyStore, signInfo)),
	}
	return serverOptions
}
