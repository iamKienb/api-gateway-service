package middleware

import (
	"context"
	"strings"

	cb "github.com/iamKienb/shopify-go-platform/circuit_breaker"

	"connectrpc.com/connect"
)

func CPInterceptor(cbManager *cb.Manager) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			parts := strings.Split(strings.TrimPrefix(req.Spec().Procedure, "/"), "/")
			serviceName := parts[0]

			breaker := cbManager.Get(serviceName)

			res, err := breaker.Execute(func() (interface{}, error) {
				return next(ctx, req)
			})

			if err != nil {
				return nil, connect.NewError(connect.CodeUnavailable, err)
			}

			return res.(connect.AnyResponse), nil
		}
	}
}
