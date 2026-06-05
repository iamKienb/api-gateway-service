package config

import (
	"time"

	configx "github.com/iamKienb/go-core/config"
)

type ApiGatewayConfig struct {
	Server                 configx.Server               `envPrefix:"API_GATEWAY"`
	CB                     configx.CircuitBreakerConfig `envPrefix:"API_GATEWAY"`
	Jwt                    configx.JwtConfig            `envPrefix:"API_GATEWAY"`
	UserCommandBaseURL     string                       `env:"API_GATEWAY_USER_COMMAND_URL" envDefault:"http://localhost:8001"`
	UserQueryBaseURL       string                       `env:"API_GATEWAY_USER_QUERY_URL" envDefault:"http://localhost:8101"`
	ShopCommandBaseURL     string                       `env:"API_GATEWAY_SHOP_COMMAND_URL" envDefault:"http://localhost:8002"`
	ShopQueryBaseURL       string                       `env:"API_GATEWAY_SHOP_QUERY_URL" envDefault:"http://localhost:8102"`
	ProductCommandBaseURL  string                       `env:"API_GATEWAY_PRODUCT_COMMAND_URL" envDefault:"http://localhost:8003"`
	ProductQueryBaseURL    string                       `env:"API_GATEWAY_PRODUCT_QUERY_URL" envDefault:"http://localhost:8103"`
	InventoryQueryBaseURL  string                       `env:"API_GATEWAY_INVENTORY_QUERY_URL" envDefault:"http://localhost:8104"`
	OrderCommandBaseURL    string                       `env:"API_GATEWAY_ORDER_COMMAND_URL" envDefault:"http://localhost:8005"`
	OrderQueryBaseURL      string                       `env:"API_GATEWAY_ORDER_QUERY_URL" envDefault:"http://localhost:8105"`
	InternalRequestTimeout time.Duration                `env:"API_GATEWAY_INTERNAL_REQUEST_TIMEOUT" envDefault:"5s"`
}
