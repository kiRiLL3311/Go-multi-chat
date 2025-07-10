package controllers

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kiRiLL3311/Go-multi-chat/initializers"
	"github.com/kiRiLL3311/Go-multi-chat/models"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	// Get the username/pass off req body
	var body struct {
		Username string `form:"username" json:"username"`
		Password string `form:"password" json:"password"`
	}

	// Bind based on content type
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if c.BindJSON(&body) != nil {
			c.Status(http.StatusBadRequest)
			slog.Error("Signup: Failed to read JSON data(body)")
			return
		}
	} else {
		if c.Bind(&body) != nil {
			c.Status(http.StatusBadRequest)
			slog.Error("Signup: Failed to read Form data(body)")
			return
		}
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Error("Signup: Failed to hash the password")
		return
	}

	// Create the user
	user := models.User{Username: body.Username, Password: string(hash)}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.Status(http.StatusBadRequest)
		slog.Warn("Signup: Username alreay exists")
		return
	}

	// For message of success to displayed
	c.HTML(http.StatusOK, "signup.html", gin.H{
		"success": true,
	})

}
func Login(c *gin.Context) {
	// Get the username and password off req body
	var body struct {
		Username string `form:"username" json:"username"`
		Password string `form:"password" json:"password"`
	}

	// Bind based on content type
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if c.BindJSON(&body) != nil {
			c.Status(http.StatusBadRequest)
			slog.Error("Login: Failed to read JSON data(body)")
			return
		}
	} else {
		if c.Bind(&body) != nil {
			c.Status(http.StatusBadRequest)
			slog.Error("Login: Failed to read Form data(body)")
			return
		}
	}

	// Look up req user
	var user models.User
	initializers.DB.First(&user, "username = ?", body.Username)

	if user.ID == 0 {
		c.Status(http.StatusBadRequest)
		slog.Info("Login: Invalid username or password")
		return
	}

	// Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Info("Login: Invalid password")
		return
	}

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 15).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.Status(http.StatusBadRequest)
		slog.Error("Login: Failed to create a token")
		return
	}

	// Set cookie and redirect to protected page
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*15, "", "", false, true)

	c.Redirect(http.StatusFound, "/chat")
}

func ChatPage(c *gin.Context) {
	// Get user from context
	userValue, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Type assert to pointer (*models.User)
	userPtr, ok := userValue.(*models.User)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "chat.html", gin.H{
		"Username": userPtr.Username,
	})
}
