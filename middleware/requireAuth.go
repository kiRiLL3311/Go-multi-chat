package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kiRiLL3311/Go-multi-chat/initializers"
	"github.com/kiRiLL3311/Go-multi-chat/models"
	"github.com/kiRiLL3311/Go-multi-chat/myLog"
)

func RequireAuth(c *gin.Context) {
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		slog.Warn("Authorization cookie missing")
		return
	}

	typeL := myLog.LogLevel{
		WARNING: "W",
		ERROR:   "E",
	}

	// Get Username
	userName, usernameErr := getUsernameFromJWT(tokenString)
	if usernameErr != "" {
		userName = "unknown"
		myLog.MidLog(userName, usernameErr, typeL.ERROR)
	}

	// Parse
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			alg := token.Header["alg"]

			// Format the error
			err := fmt.Errorf("unexpected signing method: %v", alg)

			// Log it using slog
			slog.Error("JWT signing method error",
				slog.String("error", err.Error()),
				slog.Any("alg", alg))

			return nil, err
		}
		return []byte(os.Getenv("SECRET")), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			c.AbortWithStatus(http.StatusUnauthorized)
			myLog.MidLog(userName, "Authentication failed: JWT token expired", typeL.WARNING)
			return
		default:
			c.AbortWithStatus(http.StatusUnauthorized)
			myLog.MidLog(userName, "Authentication failed: Invalid Token", typeL.ERROR)
			return
		}
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		myLog.MidLog(userName, "Authentication failed: Invalid token claims", typeL.WARNING)
		return
	}

	// Load user from DB
	var user models.User
	initializers.DB.First(&user, claims["sub"])
	if user.ID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		myLog.MidLog(userName, "Authentication failed: User not found", typeL.WARNING)
		return
	}

	// Add user to context
	c.Set("user", &user)
	// Next handler
	c.Next()
}

// Gets the authenticated user from context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	// Type-assert to *models.User
	userPtr, ok := user.(*models.User)
	return userPtr, ok
}

// Get username
func getUsernameFromJWT(tokenStr string) (string, string) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) < 2 {
		return "", "Raw username: Invalid token format"
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err.Error()
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err.Error()
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", "Raw username: Username not found in token"
	}
	return username, ""
}
