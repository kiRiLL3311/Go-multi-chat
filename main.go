package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kiRiLL3311/Go-multi-chat/controllers"
	"github.com/kiRiLL3311/Go-multi-chat/initializers"
	"github.com/kiRiLL3311/Go-multi-chat/middleware"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	router := gin.Default()

	templatePath := "./templates"

	// Load templates
	router.LoadHTMLGlob(filepath.Join(templatePath, "*.html"))

	// Web page routes
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"signupSuccess": c.Query("signup") == "success",
		})
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", gin.H{})
	})

	// API endpoints
	router.POST("/signup", controllers.Signup)
	router.POST("/login", controllers.Login)

	// Redirect root to login
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	// Required auth
	protected := router.Group("/")
	protected.Use(middleware.RequireAuth)
	{

	}

}
