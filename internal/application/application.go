package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/Powdersumm/Yandexlmsfinalproject/database"
	"github.com/Powdersumm/Yandexlmsfinalproject/handlers"
	"github.com/Powdersumm/Yandexlmsfinalproject/middleware"
	"github.com/Powdersumm/Yandexlmsfinalproject/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Request – структура входящего запроса с выражением
type Request struct {
	Expression string `json:"expression"`
}

var expressionsMutex = &sync.Mutex{}

// Expression – структура для хранения выражения и его состояния
type Expression struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID     uint   `gorm:"index"`
	Expression string
	Status     string
	Result     float64
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

// Task – структура задачи для вычисления
type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
	Expression    string  `json:"expression"`
}

// Глобальные переменные для хранения выражений и очереди задач
var expressions = make(map[string]*Expression)
var tasks = make(chan Task, 10) // Буферизованный канал для задач

// Config – конфигурация приложения
type Config struct {
	Addr string
}

// ConfigFromEnv – загрузка конфигурации из переменных окружения
func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

// Application – основная структура приложения
type Application struct {
	config *Config
}

// New – создание нового экземпляра приложения
func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

// generateUniqueID – генерация уникального идентификатора
func generateUniqueID() string {
	return uuid.New().String()
}

// parseExpression – функция для парсинга математического выражения в формате "<number> <operator> <number>"
func parseComplexExpression(expr string) (float64, error) {
	ev, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return 0, fmt.Errorf("ошибка при парсинге выражения: %v", err)
	}
	result, err := ev.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("ошибка при вычислении: %v", err)
	}
	return result.(float64), nil
}

// AddExpressionHandler – обработчик POST-запроса для добавления нового выражения
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	// Декодируем тело запроса
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid expression payload", http.StatusBadRequest)
		return
	}

	// Проверка на пустое выражение
	if req.Expression == "" {
		http.Error(w, "expression cannot be empty", http.StatusBadRequest)
		return
	}

	// Получаем userID из контекста
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Генерация ID

	// Сохраняем в БД
	newExpression := models.Expression{
		ID:         expressionID,
		UserID:     userID,
		Expression: req.Expression,
		Status:     "pending",
		Result:     0, // Изначально результат 0
	}

	if err := database.DB.Create(&newExpression).Error; err != nil {
		http.Error(w, "failed to save expression", http.StatusInternalServerError)
		return
	}

	// Отправляем задачу в канал для обработки
	tasks <- Task{
		ID:         expressionID,
		Expression: req.Expression, // Добавьте это поле в структуру Task
	}

	// Возвращаем ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expressionID})
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	var expressionList []Expression
	for _, expr := range expressions {
		expressionList = append(expressionList, *expr)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressionList,
	})
}

func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	expr, found := expressions[id]
	if !found {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expr)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, found := getNextTaskToProcess()
	if !found {
		http.Error(w, "no task available", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// Логика обработки задач
func getNextTaskToProcess() (Task, bool) {
	select {
	case task := <-tasks:
		return task, true
	default:
		return Task{}, false
	}
}

// Функция для выполнения вычислений
func processTask(task models.ExpressionTask) {
	result, err := handlers.EvaluateExpression(task.Expression)
	status := "completed"
	if err != nil {
		status = "error"
		log.Printf("Calculation error: %v", err)
	}

	// Обновление статуса в БД
	if err := database.DB.Model(&models.Expression{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{"status": status, "result": result}).Error; err != nil {
		log.Printf("DB update failed: %v", err)
	}
}

// Запуск агента для обработки задач

func startAgent() {
	for {
		select {
		case task := <-handlers.TaskQueue:
			processTask(task)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// Функция запуска приложения
func (a *Application) RunServer() error {
	// Инициализация
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	if err := database.Connect(); err != nil {
		return fmt.Errorf("DB connection failed: %v", err)
	}

	// Роутер
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			next.ServeHTTP(w, r)
		})
	})

	// Публичные эндпоинты
	r.HandleFunc("/api/v1/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.Login).Methods("POST")

	// Защищенные эндпоинты
	authRouter := r.PathPrefix("/api/v1").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)
	{
		authRouter.HandleFunc("/calculate", handlers.AddExpressionHandler).Methods("POST")
		authRouter.HandleFunc("/expressions", handlers.GetExpressionsHandler).Methods("GET")
	}

	// Запуск агента
	go startAgent()

	// Старт сервера
	log.Printf("Server started on :%s", a.config.Addr)
	return http.ListenAndServe(":"+a.config.Addr, r)
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: .env файл не найден, загрузка переменных из системы")
	}
}
