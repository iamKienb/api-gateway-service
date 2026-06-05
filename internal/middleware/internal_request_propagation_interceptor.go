package middleware

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	authx "github.com/iamKienb/go-core/middleware/auth"
)

func InternalRequestPropagationInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if reqID := authx.GetRequestID(ctx); reqID != "" {
				req.Header().Set(authx.HeaderRequestID, reqID)
			}

			if claims := authx.GetUserInfoFromCtx(ctx); claims != nil {
				if claims.UserID != "" {
					req.Header().Set(authx.HeaderUserID, claims.UserID)
				}
				if claims.Email != "" {
					req.Header().Set(authx.HeaderUserEmail, claims.Email)
				}
				if len(claims.Roles) > 0 {
					req.Header().Set(authx.HeaderUserRole, strings.Join(claims.Roles, ","))
				}
			}

			return next(ctx, req)
		}
	}
}
