звлекаем и проверяем subject (user ID)
		subClaim, ok := claims["sub"]
		if !ok {
			http.Error(w, "Missing subject claim", http.StatusUnauthorized)
			return
		}

		userID, ok := subClaim.(float64)
		if !ok {
			http.Error(w, "Invalid subject claim format", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст как uint
		ctx := context.WithValue(r.Context(), "userID", uint(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
