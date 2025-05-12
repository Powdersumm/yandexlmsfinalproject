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

const UserIDKey contextKey = "userID" // 1. Делаем ключ публичным

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Println("JWT_SECRET not set in environment")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // 2. Улучшенная проверка префикса
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

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

		if !token.Valid {
			log.Println("Invalid token validation")
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Failed to parse token claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// 3. Проверка expiration с приведением типа
		exp, ok := claims["exp"].(float64)
		if !ok {
			log.Println("Invalid expiration claim type")
			http.Error(w, "Token expiration claim is invalid", http.StatusUnauthorized)
			return
		}

		if time.Now().Unix() > int64(exp) {
			log.Println("Token expired")
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// 4. Проверка subject claim
		sub, ok := claims["sub"].(float64)
		if !ok {
			log.Printf("Invalid sub claim type: %T", claims["sub"])
			http.Error(w, "Invalid subject claim", http.StatusUnauthorized)
			return
		}

		// 5. Безопасное преобразование float64 -> uint
		if sub != float64(uint(sub)) || sub < 0 {
			log.Printf("Invalid user ID value: %f", sub)
			http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
			return
		}

		userID := uint(sub)
		log.Printf("Authenticated user ID: %d", userID) // 6. Логирование успешной аутентификации

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
