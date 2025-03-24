package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/iamgak/go-task/models"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func secureHeaders() gin.HandlerFunc {
	return (func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		c.Header("Referrer-Policy", "origin-when-cross-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "deny")
		c.Header("X-XSS-Protection", "0")
		c.Next()
	})
}
func (app *Application) LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Access Denied")
			app.Logger.Warning("Request without AuthHeader: ", authHeader)
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			app.sendJSONResponse(c.Writer, http.StatusInternalServerError, "Invalid Input Auth Header")
			app.Logger.Warning("Invalid Auth Header: ", tokenString)
			c.Abort()
			return
		}

		err := godotenv.Load()
		if err != nil {
			app.sendJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
			app.Logger.Errorf("Error fetching info from env: %T ", err)
			c.Abort()
			return
		}

		// Parse the token
		SIGNING_KEY := os.Getenv("SIGNING_KEY")
		if SIGNING_KEY == "" {
			app.sendJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
			app.Logger.Error("Error fetching info from env: ", SIGNING_KEY)
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &models.MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("[error] Unexpected signing method: %v", token.Header["alg"])
			}

			// Return the secret key
			return []byte(SIGNING_KEY), nil
		})

		if err != nil {
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Invalid Token")
			app.Logger.Error("Error fetching info from token:", err)
			c.Abort()
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(*models.MyCustomClaims); ok && token.Valid {
			// Set the username in the request context
			app.UserID = claims.UserID
			app.Email = claims.Email
			app.isAuthenticated = true
			c.Next()
		} else {
			app.Logger.Warning("Invalid Token")
			app.sendJSONResponse(c.Writer, http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
	}
}

// func (app *Application) rateLimit() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var (
// 			mu      sync.Mutex
// 			clients = make(map[string]*rate.Limiter)
// 		)
// 		ip, _, err := net.SplitHostPort(c.RemoteIP())
// 		if err != nil {
// 			app.Logger.Errorf("Internal Error: %T", err)
// 			app.sendJSONResponse(c.Writer, 500, "Internal Server Error")
// 			return
// 		}
// 		mu.Lock()
// 		if _, found := clients[ip]; !found {
// 			clients[ip] = rate.NewLimiter(2, 4)
// 		}
// 		if !clients[ip].Allow() {
// 			mu.Unlock()
// 			app.Logger.Error("Too many Requests")
// 			app.sendJSONResponse(c.Writer, 500, "Internal Server Error")
// 			return
// 		}
// 		mu.Unlock()
// 	}
// }

func (app *Application) rateLimiter() gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			// it will run until code run but take break every minute laziness
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()
			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.ClientIP())
		if err != nil {
			ip := net.ParseIP(c.ClientIP())
			if ip == nil {
				app.Logger.Warn("Invalid IP address :", c.ClientIP())
				app.ServerError(c.Writer, err)
				return
			}
		}
		// Lock the mutex to prevent this code from being executed concurrently.

		mu.Lock()
		if _, found := clients[ip]; !found {
			// Create and add a new client struct to the map if it doesn't already exist.
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(5), 3),
			}
		}

		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.CustomError(c.Writer, http.StatusTooManyRequests, "Too, many request. Rate Limit Exceed")
			return
		}

		mu.Unlock()
		// c.Next()
	}
}

func (app *Application) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the original request context with the new one
		c.Request = c.Request.WithContext(ctx)

		// Continue to the next handler
		c.Next()

		// If the context was canceled, return timeout error
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
			c.Abort()
			return
		}
	}
}

func MaintenanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetHeader("X-User-Role") // only open for me

		if os.Getenv("SERVER_STATUS") == "maintenance" && userRole != "admin" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "The server is currently under maintenance. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
