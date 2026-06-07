package module

import (
	"api-gateway/internal/bootstrap/config"
	"api-gateway/internal/middleware"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/iamKienb/api-contract/gen/category/categoryconnect"
	"github.com/iamKienb/api-contract/gen/inventory/inventoryconnect"
	"github.com/iamKienb/api-contract/gen/order/orderconnect"
	"github.com/iamKienb/api-contract/gen/otp/otpconnect"
	"github.com/iamKienb/api-contract/gen/product/productconnect"
	"github.com/iamKienb/api-contract/gen/shop/shopconnect"
	"github.com/iamKienb/api-contract/gen/user/userconnect"
	cb "github.com/iamKienb/go-core/circuit_breaker"
	jwtx "github.com/iamKienb/go-core/jwt"
	authx "github.com/iamKienb/go-core/middleware/auth"
	observabilityx "github.com/iamKienb/go-core/middleware/observability"
)

type InternalServiceClient struct {
	UserCommand    userconnect.UserCommandServiceClient
	UserQuery      userconnect.UserQueryServiceClient
	OtpCommand     otpconnect.OTPCommandServiceClient
	ShopCommand    shopconnect.ShopCommandServiceClient
	ShopQuery      shopconnect.ShopQueryServiceClient
	ProductCommand productconnect.ProductCommandServiceClient
	ProductQuery   productconnect.ProductQueryServiceClient
	Category       categoryconnect.CategoryCommandServiceClient
	InventoryQuery inventoryconnect.InventoryQueryServiceClient
	OrderCommand   orderconnect.OrderCommandClient
	OrderQuery     orderconnect.OrderQueryClient
}

func NewInternalServiceClient(httpClient connect.HTTPClient, cfg *config.ApiGatewayConfig, opts ...connect.ClientOption) *InternalServiceClient {
	return &InternalServiceClient{
		UserCommand:    userconnect.NewUserCommandServiceClient(httpClient, cfg.UserCommandBaseURL, opts...),
		UserQuery:      userconnect.NewUserQueryServiceClient(httpClient, cfg.UserQueryBaseURL, opts...),
		OtpCommand:     otpconnect.NewOTPCommandServiceClient(httpClient, cfg.UserCommandBaseURL, opts...),
		ShopCommand:    shopconnect.NewShopCommandServiceClient(httpClient, cfg.ShopCommandBaseURL, opts...),
		ShopQuery:      shopconnect.NewShopQueryServiceClient(httpClient, cfg.ShopQueryBaseURL, opts...),
		ProductCommand: productconnect.NewProductCommandServiceClient(httpClient, cfg.ProductCommandBaseURL, opts...),
		ProductQuery:   productconnect.NewProductQueryServiceClient(httpClient, cfg.ProductQueryBaseURL, opts...),
		Category:       categoryconnect.NewCategoryCommandServiceClient(httpClient, cfg.ProductCommandBaseURL, opts...),
		InventoryQuery: inventoryconnect.NewInventoryQueryServiceClient(httpClient, cfg.InventoryQueryBaseURL, opts...),
		OrderCommand:   orderconnect.NewOrderCommandClient(httpClient, cfg.OrderCommandBaseURL, opts...),
		OrderQuery:     orderconnect.NewOrderQueryClient(httpClient, cfg.OrderQueryBaseURL, opts...),
	}
}

type AdapterModule struct {
	Mux *http.ServeMux
}

func NewAdapterModule(logger *slog.Logger, cfg *config.ApiGatewayConfig) (*AdapterModule, error) {
	jwtService, err := jwtx.NewVerifier(cfg.Jwt)
	if err != nil {
		return nil, fmt.Errorf("jwt verifier: %w", err)
	}

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

	internalHTTPClient := &http.Client{Timeout: cfg.InternalRequestTimeout}
	internalSvc := NewInternalServiceClient(internalHTTPClient, cfg, clientOpts)

	mux.Handle(userconnect.NewUserCommandServiceHandler(internalSvc.UserCommand, serverOpts))
	mux.Handle(userconnect.NewUserQueryServiceHandler(internalSvc.UserQuery, serverOpts))
	mux.Handle(otpconnect.NewOTPCommandServiceHandler(internalSvc.OtpCommand, serverOpts))
	mux.Handle(shopconnect.NewShopCommandServiceHandler(internalSvc.ShopCommand, serverOpts))
	mux.Handle(shopconnect.NewShopQueryServiceHandler(internalSvc.ShopQuery, serverOpts))
	mux.Handle(productconnect.NewProductCommandServiceHandler(internalSvc.ProductCommand, serverOpts))
	mux.Handle(productconnect.NewProductQueryServiceHandler(internalSvc.ProductQuery, serverOpts))
	mux.Handle(categoryconnect.NewCategoryCommandServiceHandler(internalSvc.Category, serverOpts))
	mux.Handle(inventoryconnect.NewInventoryQueryServiceHandler(internalSvc.InventoryQuery, serverOpts))
	mux.Handle(orderconnect.NewOrderCommandHandler(internalSvc.OrderCommand, serverOpts))
	mux.Handle(orderconnect.NewOrderQueryHandler(internalSvc.OrderQuery, serverOpts))

	reflector := grpcreflect.NewStaticReflector(
		userconnect.UserCommandServiceName,
		userconnect.UserQueryServiceName,
		otpconnect.OTPCommandServiceName,
		shopconnect.ShopCommandServiceName,
		shopconnect.ShopQueryServiceName,
		productconnect.ProductCommandServiceName,
		productconnect.ProductQueryServiceName,
		categoryconnect.CategoryCommandServiceName,
		inventoryconnect.InventoryQueryServiceName,
		orderconnect.OrderCommandName,
		orderconnect.OrderQueryName,
	)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return &AdapterModule{
		Mux: mux,
	}, nil
}
