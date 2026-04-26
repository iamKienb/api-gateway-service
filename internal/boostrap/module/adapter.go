package module

import (
	"log/slog"
	"net/http"
	"shopify-api-gateway/internal/boostrap/config"
	"shopify-api-gateway/internal/middleware"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/iamKienb/shopify-go-api/gen/otp/otpconnect"
	"github.com/iamKienb/shopify-go-api/gen/user/userconnect"
	cb "github.com/iamKienb/shopify-go-platform/circuit_breaker"
	authx "github.com/iamKienb/shopify-go-platform/middleware/auth"
	observabilityx "github.com/iamKienb/shopify-go-platform/middleware/observability"
)

type InternalServiceClient struct {
	User userconnect.UserCommandServiceClient
	Otp  otpconnect.OTPCommandServiceClient
}

func NewInternalServiceClient(httpClient connect.HTTPClient, opts ...connect.ClientOption) *InternalServiceClient {
	return &InternalServiceClient{
		User: userconnect.NewUserCommandServiceClient(httpClient, "http://localhost:8888", opts...),
		Otp:  otpconnect.NewOTPCommandServiceClient(httpClient, "http://localhost:8888", opts...),
	}
}

type AdapterModule struct {
	Mux *http.ServeMux
}

func NewAdapterModule(logger *slog.Logger, cfg *config.ApiGatewayConfig) *AdapterModule {
	authJwt := authx.NewJWTGenerator(cfg.Jwt)
	cbManager := cb.NewManager(cfg.CB)

	var interceptors []connect.Interceptor

	tracingInterceptor, err := observabilityx.TracingInterceptor()
	if err != nil {
		logger.Error("failed to initialize tracing interceptor", slog.Any("error", err))
	} else {
		interceptors = append(interceptors, tracingInterceptor)
	}

	interceptors = append(interceptors,
		observabilityx.RecoveryInterceptor(logger),
		observabilityx.LoggingInterceptor(logger),
		middleware.AuthInterceptor(authJwt),
		middleware.CPInterceptor(cbManager),
		observabilityx.ErrorResponseInterceptor(),
	)

	allInterceptors := connect.WithInterceptors(interceptors...)
	mux := http.NewServeMux()

	internalSvc := NewInternalServiceClient(http.DefaultClient, allInterceptors)

	mux.Handle(userconnect.NewUserCommandServiceHandler(internalSvc.User))
	mux.Handle(otpconnect.NewOTPCommandServiceHandler(internalSvc.Otp))

	reflector := grpcreflect.NewStaticReflector(
		userconnect.UserCommandServiceName,
		otpconnect.OTPCommandServiceName,
	)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return &AdapterModule{
		Mux: mux,
	}
}
