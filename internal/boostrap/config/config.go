package config

import (
	configx "github.com/iamKienb/shopify-go-platform/config"
)

type ApiGatewayConfig struct {
	Server configx.Server               `envPrefix:"API_GATEWAY"`
	CB     configx.CircuitBreakerConfig `envPrefix:"API_GATEWAY"`
	Jwt    configx.JwtConfig            `envPrefix:"API_GATEWAY"`
}
