package controllers

import (
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
			return
		}
	} else {
		if c.Bind(&body) != nil {
			c.HTML(http.StatusBadRequest, "signup.html", gin.H{"error": "Failed to read form data"})
			return
		}
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.HTML(http.StatusBadRequest, "signup.html", gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user
	user := models.User{Username: body.Username, Password: string(hash)}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.HTML(http.StatusBadRequest, "signup.html", gin.H{"error": "Username already exists"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
			return
		}
	} else {
		if c.Bind(&body) != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Failed to read form data"})
			return
		}
	}

	// Look up req user
	var user models.User
	initializers.DB.First(&user, "username = ?", body.Username)

	if user.ID == 0 {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Invalid username or password"})
		return
	}

	// Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Invalid username or password"})
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
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Failed to create token"})
		return
	}

	// Set cookie and redirect to protected page
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*5, "", "", false, true)

	c.Redirect(http.StatusFound, "/chat")
}
