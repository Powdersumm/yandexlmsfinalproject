package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Powdersumm/Yandexlmsfinalproject/database"
	"github.com/Powdersumm/Yandexlmsfinalproject/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register - обработчик регистрации
func Register(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка существования пользователя
	var user models.User
	database.DB.Where("login = ?", req.Login).First(&user)
	if user.ID != 0 {
		http.Error(w, "User exists", http.StatusConflict)
		return
	}

	// Хеширование пароля
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	newUser := models.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}
	database.DB.Create(&newUser)

	// Ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created"})
}

// Login - обработчик входа
func Login(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Поиск пользователя
	var user models.User
	database.DB.Where("login = ?", req.Login).First(&user)
	if user.ID == 0 {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	// Ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": tokenString,
		"token_type":   "bearer",
	})
}

// AddExpressionHandler - обработчик для добавления нового выражения
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Expression string `json:"expression"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// (Опционально) Проверяем, что выражение не пустое
	if req.Expression == "" {
		http.Error(w, "Expression cannot be empty", http.StatusBadRequest)
		return
	}

	// Генерация ID выражения (можно использовать UUID)
	expressionID := time.Now().UnixNano()

	// Отправляем JSON-ответ с ID выражения
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         expressionID,
		"expression": req.Expression,
		"status":     "pending",
	})
}

// GetExpressionsHandler - обработчик для получения списка выражений
func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	expressions := []map[string]interface{}{
		{"id": 1, "expression": "2 + 2", "result": 4, "status": "completed"},
		{"id": 2, "expression": "5 * 5", "result": 25, "status": "completed"},
	}

	// Отправляем JSON с примерами выражений
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}
