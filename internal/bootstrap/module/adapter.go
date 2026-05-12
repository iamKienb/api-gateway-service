package module

import (
	"api-gateway-module/internal/bootstrap/config"
	"api-gateway-module/internal/middleware"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/iamKienb/shopify-go-api/gen/otp/otpconnect"
	"github.com/iamKienb/shopify-go-api/gen/user/userconnect"
	cb "github.com/iamKienb/shopify-go-platform/circuit_breaker"
	jwtx "github.com/iamKienb/shopify-go-platform/jwt"
	authx "github.com/iamKienb/shopify-go-platform/middleware/auth"
	observabilityx "github.com/iamKienb/shopify-go-platform/middleware/observability"
)

type InternalServiceClient struct {
	User userconnect.UserCommandServiceClient
	Otp  otpconnect.OTPCommandServiceClient
}

func NewInternalServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) *InternalServiceClient {
	return &InternalServiceClient{
		User: userconnect.NewUserCommandServiceClient(httpClient, baseURL, opts...),
		Otp:  otpconnect.NewOTPCommandServiceClient(httpClient, baseURL, opts...),
	}
}

type AdapterModule struct {
	Mux *http.ServeMux
}

func NewAdapterModule(logger *slog.Logger, cfg *config.ApiGatewayConfig) *AdapterModule {
	jwtService, _ := jwtx.New(cfg.Jwt)
	cbManager := cb.NewManager(cfg.CB)

	var serverInterceptors []connect.Interceptor
	var clientInterceptors []connect.Interceptor

	tracingInterceptor, err := observabilityx.TracingInterceptor()
	if err != nil {
		logger.Error("failed to initialize tracing interceptor", slog.Any("error", err))
	} else {
		serverInterceptors = append(serverInterceptors, tracingInterceptor)
		clientInterceptors = append(clientInterceptors, tracingInterceptor)
	}

	serverInterceptors = append(serverInterceptors,
		observabilityx.RecoveryInterceptor(logger),
		authx.RequestContextInterceptor(),
		observabilityx.LoggingInterceptor(logger),
		middleware.AuthInterceptor(jwtService),
		observabilityx.ErrorResponseInterceptor(logger),
	)

	clientInterceptors = append(clientInterceptors,
		middleware.InternalRequestPropagationInterceptor(),
		middleware.CircuitBreakerInterceptor(cbManager),
		observabilityx.LoggingInterceptor(logger),
		observabilityx.ErrorResponseInterceptor(logger),
	)

	serverOpts := connect.WithInterceptors(serverInterceptors...)
	clientOpts := connect.WithInterceptors(clientInterceptors...)
	mux := http.NewServeMux()

	internalHTTPClient := &http.Client{Timeout: cfg.UserCommandTimeout}
	internalSvc := NewInternalServiceClient(internalHTTPClient, cfg.UserCommandBaseURL, clientOpts)

	mux.Handle(userconnect.NewUserCommandServiceHandler(internalSvc.User, serverOpts))
	mux.Handle(otpconnect.NewOTPCommandServiceHandler(internalSvc.Otp, serverOpts))

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
