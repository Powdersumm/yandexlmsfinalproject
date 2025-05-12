package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/Powdersumm/Yandexlmsfinalproject/database"
	"github.com/Powdersumm/Yandexlmsfinalproject/middleware"
	"github.com/Powdersumm/Yandexlmsfinalproject/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register - обработчик регистрации с улучшенной обработкой ошибок
func Register(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Валидация
	if len(req.Login) < 3 || len(req.Password) < 6 {
		http.Error(w, "Login (3+ chars) and password (6+ chars) required", http.StatusBadRequest)
		return
	}

	// Проверка существующего пользователя
	var existingUser models.User
	if database.DB.Where("login = ?", req.Login).First(&existingUser).Error == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Password processing failed", http.StatusInternalServerError)
		return
	}

	// Создание пользователя
	newUser := models.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}
	if err := database.DB.Create(&newUser).Error; err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created"})
}

// Login - обработчик входа с улучшенной безопасностью
func Login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Поиск пользователя
	var user models.User
	if database.DB.Where("login = ?", req.Login).First(&user).Error != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерация JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Ответ
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": tokenString,
		"token_type":   "bearer",
		"expires_in":   int(time.Until(expirationTime).Seconds()),
	})
}

// AddExpressionHandler - обработчик выражений с сохранением в БД
var TaskQueue = make(chan models.ExpressionTask, 100)

// Добавлено: регулярное выражение для валидации выражений
var validExpressionRegex = regexp.MustCompile(`^[\d\s+\-*/()]+$`)

func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Expression string `json:"expression"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Expression == "" || !validExpressionRegex.MatchString(req.Expression) {
		http.Error(w, "Invalid expression format", http.StatusBadRequest)
		return
	}

	// UserID из контекста
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Сохранение в БД
	expressionID := uuid.New().String()
	expr := models.Expression{
		ID:         expressionID,
		UserID:     userID,
		Expression: req.Expression,
		Status:     "pending",
	}
	if err := database.DB.Create(&expr).Error; err != nil {
		http.Error(w, "Failed to save expression", http.StatusInternalServerError)
		return
	}

	// Отправка задачи в очередь
	TaskQueue <- models.ExpressionTask{
		ID:         expressionID,
		Expression: req.Expression,
		UserID:     userID,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expressionID})
}

// GetExpressionsHandler - получение выражений из БД
func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Пагинация
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 10
	if l, _ := strconv.Atoi(r.URL.Query().Get("limit")); l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit

	// Запрос к БД
	var expressions []models.Expression
	query := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)
	if err := query.Find(&expressions).Error; err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	response := make([]map[string]interface{}, len(expressions))
	for i, expr := range expressions {
		response[i] = map[string]interface{}{
			"id":         expr.ID,
			"expression": expr.Expression,
			"status":     expr.Status,
			"result":     expr.Result,
			"created_at": expr.CreatedAt,
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": response})
}

func EvaluateExpression(expr string) (float64, error) {
	expression, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return 0, fmt.Errorf("ошибка парсинга выражения: %v", err)
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("ошибка вычисления: %v", err)
	}

	return result.(float64), nil
}
