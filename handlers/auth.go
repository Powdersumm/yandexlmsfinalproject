package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/Powdersumm/Yandexlmscalcproject2sprint/database"
	"github.com/Powdersumm/Yandexlmscalcproject2sprint/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверка существования пользователя
	var user models.User
	database.DB.Where("login = ?", req.Login).First(&user)
	if user.ID != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User exists"})
		return
	}

	// Хеширование пароля
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	newUser := models.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}
	database.DB.Create(&newUser)

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

func Login(c *gin.Context) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Поиск пользователя
	var user models.User
	database.DB.Where("login = ?", req.Login).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
		"token_type":   "bearer",
	})
}
