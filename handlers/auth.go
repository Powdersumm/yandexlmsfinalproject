package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/Powdersumm/Yandexlmsfinalproject/database"
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

	// Валидация входных данных
	if len(req.Login) < 3 || len(req.Password) < 6 {
		http.Error(w, "Login must be at least 3 characters, password 6 characters", http.StatusBadRequest)
		return
	}

	var existingUser models.User
	if err := database.DB.Where("login = ?", req.Login).First(&existingUser).Error; err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

// Login - обработчик входа с улучшенной безопасностью
func Login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Поиск пользователя в базе данных
	var user models.User
	if err := database.DB.Where("login = ?", req.Login).First(&user).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерация JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	expiresIn := int(time.Until(expirationTime).Seconds())

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": tokenString,
		"token_type":   "bearer",
		"expires_in":   expiresIn,
		"user_id":      user.ID,
	})
}

// AddExpressionHandler - обработчик выражений с сохранением в БД
var TaskQueue = make(chan models.ExpressionTask, 100)

// Добавлено: регулярное выражение для валидации выражений
var validExpressionRegex = regexp.MustCompile(`^[\d\s+\-*/()]+$`)

// AddExpressionHandler - улучшенная версия с очередью и валидацией
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Expression string `json:"expression"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация выражения
	if req.Expression == "" {
		http.Error(w, "Expression cannot be empty", http.StatusBadRequest)
		return
	}

	if !validExpressionRegex.MatchString(req.Expression) {
		http.Error(w, "Expression contains invalid characters. Only numbers, +-*/() and spaces allowed",
			http.StatusBadRequest)
		return
	}

	// Получаем userID из контекста
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Генерация UUID
	expressionID := uuid.New().String()

	newExpression := models.Expression{
		ID:         expressionID,
		UserID:     userID,
		Expression: req.Expression,
		Status:     "pending",
		Result:     0,
	}

	// Сохранение в БД
	if err := database.DB.Create(&newExpression).Error; err != nil {
		http.Error(w, "Failed to save expression: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправка задачи в очередь
	task := models.ExpressionTask{
		ID:         expressionID,
		Expression: req.Expression,
		UserID:     userID,
	}

	select {
	case TaskQueue <- task:
		// Задача успешно добавлена в очередь
	default:
		http.Error(w, "Server is too busy. Please try again later", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         expressionID,
		"expression": req.Expression,
		"status":     "pending",
	})
}

// GetExpressionsHandler - получение выражений из БД
func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Параметры пагинации
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var expressions []models.Expression
	query := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	if err := query.Find(&expressions).Error; err != nil {
		http.Error(w, "Failed to retrieve expressions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение общего количества для пагинации
	var totalCount int64
	if err := database.DB.Model(&models.Expression{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		http.Error(w, "Failed to get total count", http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": response,
		"pagination": map[string]interface{}{
			"total":     totalCount,
			"page":      page,
			"limit":     limit,
			"last_page": (totalCount + int64(limit) - 1) / int64(limit),
		},
	})
}
