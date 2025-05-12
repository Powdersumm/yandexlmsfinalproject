package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Парсинг токена с проверкой метода подписи
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("Token parsing error: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверка валидности токена ДО извлечения claims
		if !token.Valid {
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Проверка срока действия
		exp, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(exp) {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// Проверка типа sub и конвертация
		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, "Invalid subject claim (must be numeric)", http.StatusUnauthorized)
			return
		}

		// Конвертация float64 -> uint с проверкой
		if sub != float64(uint(sub)) {
			http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, uint(sub))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
