package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Powdersumm/Yandexlmsfinalproject/database"
	"github.com/Powdersumm/Yandexlmsfinalproject/handlers"
	"github.com/Powdersumm/Yandexlmsfinalproject/middleware"
	"github.com/Powdersumm/Yandexlmsfinalproject/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Config — конфигурация приложения
type Config struct {
	Addr string
}

// ConfigFromEnv — загрузка конфигурации из переменных окружения
func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

// Application — основная структура приложения
type Application struct {
	config *Config
	db     *gorm.DB
}

// New — создание нового экземпляра приложения
func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

// generateUniqueID — генерация уникального идентификатора
func generateUniqueID() string {
	return uuid.New().String()
}

// AddExpressionHandler — обработчик POST-запроса для добавления нового выражения
func (a *Application) AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Expression == "" {
		http.Error(w, "Expression cannot be empty", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Генерация UUID
	expressionID := generateUniqueID()

	newExpression := models.Expression{
		ID:         expressionID,
		UserID:     userID,
		Expression: req.Expression,
		Status:     "pending",
	}

	if err := database.DB.Create(&newExpression).Error; err != nil {
		log.Printf("Ошибка сохранения выражения: %v", err)
		http.Error(w, "Failed to save expression", http.StatusInternalServerError)
		return
	}

	// Отправка задачи в очередь
	handlers.TaskQueue <- models.ExpressionTask{
		ID:         expressionID,
		Expression: req.Expression,
		UserID:     userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expressionID})
}

// processTask — обработка задачи
func (a *Application) processTask(task models.ExpressionTask) {
	result, err := handlers.EvaluateExpression(task.Expression)
	status := "completed"
	if err != nil {
		status = "error"
		log.Printf("Ошибка вычисления: %v", err)
	}

	// Обновление статуса в БД
	if err := database.DB.Model(&models.Expression{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"status": status,
			"result": result,
		}).Error; err != nil {
		log.Printf("Ошибка обновления БД: %v", err)
	}
}

// startAgent — запуск агента для обработки задач
func (a *Application) startAgent() {
	for {
		select {
		case task := <-handlers.TaskQueue:
			a.processTask(task)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// RunServer — запуск сервера
func (a *Application) RunServer() error {
	if err := godotenv.Load(); err != nil {
		log.Println("Не найден .env файл")
	}

	if err := database.Connect(); err != nil {
		return fmt.Errorf("Ошибка подключения к БД: %v", err)
	}

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			next.ServeHTTP(w, r)
		})
	})

	// Публичные маршруты
	r.HandleFunc("/api/v1/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.Login).Methods("POST")

	// Защищенные маршруты
	authRouter := r.PathPrefix("/api/v1").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)
	{
		authRouter.HandleFunc("/calculate", a.AddExpressionHandler).Methods("POST")
		authRouter.HandleFunc("/expressions", handlers.GetExpressionsHandler).Methods("GET")
	}

	go a.startAgent()

	log.Printf("Сервер запущен на порту %s", a.config.Addr)
	return http.ListenAndServe(":"+a.config.Addr, r)
}
