package middleware

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	app_error "github.com/iamKienb/go-core/app_error"
	jwtx "github.com/iamKienb/go-core/jwt"
	authx "github.com/iamKienb/go-core/middleware/auth"
)

var publicProcedures = map[string]bool{
	"/user.v1.UserCommandService/LoginUser":    true,
	"/user.v1.UserCommandService/RegisterUser": true,

	"/otp.v1.OTPCommandService/Verify": true,
	"/otp.v1.OTPCommandService/Resend": true,

	"/product.v1.ProductQueryService/GetProductDetail":      true,
	"/product.v1.ProductQueryService/SearchProducts":        true,
	"/product.v1.ProductQueryService/ListProductVariants":   true,
	"/product.v1.ProductQueryService/ListProductCategories": true,

	"/shop.v1.ShopQueryService/GetShopDetail": true,
	"/shop.v1.ShopQueryService/SearchShops":   true,
}

func AuthInterceptor(service jwtx.JWTXService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure
			if publicProcedures[procedure] {
				return next(ctx, req)
			}

			authHeader := req.Header().Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return nil, app_error.New(app_error.KindUnauthorized, "auth_required", "please log in first", nil)

			}

			if service == nil {
				return nil, app_error.New(app_error.KindInternal, "jwt_verifier_unavailable", "authentication service is not configured", nil)
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenStr == "" {
				return nil, app_error.New(app_error.KindUnauthorized, "auth_required", "please log in first", nil)
			}

			claims, err := service.Verify(tokenStr)
			if err != nil {
				return nil, app_error.New(app_error.KindUnauthorized, "session_expired", "your session has expired, please log in again", err)
			}
			if claims == nil || strings.TrimSpace(claims.UserID) == "" {
				return nil, app_error.New(app_error.KindUnauthorized, "invalid_session", "your session is invalid, please log in again", nil)
			}

			ctx = authx.SetUserInfoToCtx(ctx, claims)
			req.Header().Set(authx.HeaderUserID, claims.UserID)
			req.Header().Set(authx.HeaderUserEmail, claims.Email)
			req.Header().Set(authx.HeaderUserName, claims.FullName)
			req.Header().Set(authx.HeaderUserRole, strings.Join(claims.Roles, ","))

			return next(ctx, req)
		}
	}
}
