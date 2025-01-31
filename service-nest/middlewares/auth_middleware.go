package middlewares

import (
	"context"
	"github.com/golang-jwt/jwt"
	"net/http"
	"service-nest/logger"
	"service-nest/response"
	"service-nest/util"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			response.ErrorResponse(w, http.StatusUnauthorized, "Missing token", 1002)
			return
		}

		// Split Bearer from token
		tokenString := strings.Split(authorizationHeader, "Bearer ")[1]

		// Verify JWT token
		token, err := util.VerifyJWT(tokenString)
		if err != nil {
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid token", 1002)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid token", 1002)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid token", 1002)
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid token", 1002)
			return
		}

		// Store both the userID and role in the request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "role", role)

		// Pass the context with the userID and role to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HouseHolderAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)
		if role != "Householder" {
			logger.Error("Invalid role", nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid role", 1007)

			return
		}
		next.ServeHTTP(w, r)
	})
}
func ServiceProviderAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)
		if role != "ServiceProvider" {
			logger.Error("Invalid role", nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid role", 1007)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)
		if role != "Admin" {
			logger.Error("Invalid role", nil)
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid role", 1007)
			return
		}
		next.ServeHTTP(w, r)
	})
}
