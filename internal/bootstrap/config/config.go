package config

import (
	"time"

	configx "github.com/iamKienb/shopify-go-platform/config"
)

type ApiGatewayConfig struct {
	Server             configx.Server               `envPrefix:"API_GATEWAY"`
	CB                 configx.CircuitBreakerConfig `envPrefix:"API_GATEWAY"`
	Jwt                configx.JwtConfig            `envPrefix:"API_GATEWAY"`
	UserCommandBaseURL string                       `env:"API_GATEWAY_USER_COMMAND_URL" envDefault:"http://localhost:8888"`
	UserCommandTimeout time.Duration                `env:"API_GATEWAY_USER_COMMAND_TIMEOUT" envDefault:"5s"`
}
