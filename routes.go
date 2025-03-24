package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (app *Application) InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(MaintenanceMiddleware())
	r.Use(app.TimeoutMiddleware(5 * time.Second))
	// read API
	r.GET("/tasks", app.ListTask)
	r.GET("/tasks/:id", app.TaskListingById)

	authorise := r.Group("/tasks")

	authorise.Use(app.LoginMiddleware(), secureHeaders(), app.rateLimiter())
	{
		// write API
		authorise.POST("/", app.CreateTask)
		authorise.PUT("/update/:id", app.UpdateTask)
		authorise.DELETE("/delete/:id", app.SoftDelete)
	}

	r.POST("/login", app.UserLogin)
	r.POST("/register", app.UserRegister)
	r.GET("/activation_token/:token", app.UserActivateAccount)
	return r
}
