package middleware

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	authx "github.com/iamKienb/shopify-go-platform/middleware/auth"
)

var publicProcedures = map[string]bool{
	"/user.v1.UserCommandService/Login":    true,
	"/user.v1.UserCommandService/Register": true,

	"/otp.v1.OTPCommandService/Verify": true,
	"/otp.v1.OTPCommandService/Resend": true,
}

func AuthInterceptor(auth authx.Generator) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure
			if publicProcedures[procedure] {
				return next(ctx, req)
			}

			authHeader := req.Header().Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Please log in first"))

			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := auth.Verify(tokenStr)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Your session has expired. Please log in again to continue"))
			}

			req.Header().Set("X-User-Id", claims.UserID)
			req.Header().Set("X-User-Email", claims.Email)
			req.Header().Set("X-User-Roles", strings.Join(claims.Roles, ","))

			return next(ctx, req)
		}
	}
}
