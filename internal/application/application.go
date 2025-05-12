package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Powdersumm/Yandexlmsfinalproject/database"
	"github.com/Powdersumm/Yandexlmsfinalproject/handlers"
	"github.com/Powdersumm/Yandexlmsfinalproject/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Config struct {
	Addr string
}

type Application struct {
	config *Config
}

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
}

var (
	expressions   = make(map[string]*handlers.Expression)
	tasks         = make(chan Task, 10)
	expressionsMu = &sync.Mutex{}
)

func ConfigFromEnv() *Config {
	return &Config{
		Addr: getEnv("PORT", "8080"),
	}
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

func (a *Application) RunServer() error {
	if err := LoadEnv(); err != nil {
		return err
	}

	if err := database.Connect(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/v1/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.Login).Methods("POST")
	r.HandleFunc("/internal/task", GetTaskHandler).Methods("GET")

	// Protected routes
	authRouter := r.PathPrefix("/api/v1").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)
	{
		authRouter.HandleFunc("/calculate", handlers.AddExpressionHandler).Methods("POST")
		authRouter.HandleFunc("/expressions", handlers.GetExpressionsHandler).Methods("GET")
		authRouter.HandleFunc("/expressions/{id}", handlers.GetExpressionByIDHandler).Methods("GET")
	}

	go startAgent()

	log.Printf("Server started on port %s", a.config.Addr)
	return http.ListenAndServe(":"+a.config.Addr, r)
}

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		log.Print("Warning: .env file not found, using system environment variables")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-tasks:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	default:
		http.Error(w, "No tasks available", http.StatusNotFound)
	}
}

func startAgent() {
	for {
		select {
		case task := <-tasks:
			processTask(task)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func processTask(task Task) {
	// ... (ваша реализация обработки задач)
}
