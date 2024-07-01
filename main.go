// main.go
// Author: Bipin Kumar Ojha (Freelancer)

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/bkojha74/task-management/database"
	"github.com/bkojha74/task-management/handlers"
	"github.com/bkojha74/task-management/helper"
	"github.com/bkojha74/task-management/middleware"
	"github.com/bkojha74/task-management/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Read current working directory
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Load environment variables from the configuration file
	helper.LoadEnv(currentWorkDirectory + "/config")

	// Retrieve environment variables
	mongoURI := helper.GetEnv("MONGO_URI")
	appPort := helper.GetEnv("APP_PORT")
	jwtSecret := helper.GetEnv("JWT_SECRET")
	tokenExpiry := helper.GetEnv("TOKEN_EXPIRY_TIME")

	// Ensure required environment variables are set
	if mongoURI == "" || jwtSecret == "" || appPort == "" || tokenExpiry == "" {
		log.Fatal("Environment variables MONGO_URI, APP_PORT, JWT_SECRET, and TOKEN_EXPIRY_TIME must be set")
	}

	// Convert TOKEN_EXPIRY_TIME to integer
	tokenExpiryTime, err := strconv.Atoi(tokenExpiry)
	if err != nil {
		log.Fatal("Error converting TOKEN_EXPIRY_TIME to integer:", err)
	}

	// Initialize the Fiber app
	app := fiber.New()

	// Middleware setup
	app.Use(logger.New())  // Request logger middleware
	app.Use(recover.New()) // Panic recovery middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://cloud.mongodbs.com", // CORS configuration for allowed origins
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	})) // CORS middleware
	app.Use(limiter.New(limiter.Config{
		Max:        20,        // Maximum number of requests per IP
		Expiration: 60 * 1000, // Time window for rate limiting in milliseconds
	})) // Request rate limiter middleware

	// Initialize MongoDB connection
	database.Init(mongoURI)
	defer database.Disconnect() // Ensure database connection is closed when main function exits

	// User management endpoints
	app.Post("/signup", handlers.SignUp)                             // User registration endpoint
	app.Post("/signin", handlers.SignIn(jwtSecret, tokenExpiryTime)) // User login endpoint with JWT token generation
	app.Post("/signout", handlers.SignOut)                           // User logout endpoint

	// JWT Middleware for task management endpoints
	app.Use("/tasks", middleware.Protected(jwtSecret))

	// Task management endpoints
	app.Post("/tasks", utils.JWTMiddleware(jwtSecret), handlers.CreateTask)       // Create task endpoint
	app.Get("/tasks", utils.JWTMiddleware(jwtSecret), handlers.GetTasks)          // Get all tasks endpoint
	app.Get("/tasks/:id", utils.JWTMiddleware(jwtSecret), handlers.GetTask)       // Get a single task by ID endpoint
	app.Put("/tasks/:id", utils.JWTMiddleware(jwtSecret), handlers.UpdateTask)    // Update task by ID endpoint
	app.Delete("/tasks/:id", utils.JWTMiddleware(jwtSecret), handlers.DeleteTask) // Delete task by ID endpoint

	// Start the Fiber server on the specified port
	log.Fatal(app.Listen(":" + appPort))
}
