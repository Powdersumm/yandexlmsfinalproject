package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware для Gorilla Mux
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем заголовок Authorization
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Удаляем "Bearer " из строки, если токен передан в стандартном формате
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Парсим токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Проверяем валидность токена
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Добавляем userID в контекст запроса
			ctx := r.Context()
			ctx = context.WithValue(ctx, "userID", claims["sub"])

			// Передаём управление следующему обработчику с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}
	})
}
