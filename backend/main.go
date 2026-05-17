package main

import (
	"log"
	"os"

	"atomquest-portal/pkg/api" // Injects our new API router
	"atomquest-portal/pkg/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Fetch database connection string from environment context safely, fallback to default
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "user=postgres password=postgres dbname=atomquest sslmode=disable"
	}

	// 1. Fire up the relational datastore pool
	database := db.InitDB(dsn)
	defer database.Close()

	// 2. Instantiate our Fiber application engine with extreme custom tuning parameters
	app := fiber.New(fiber.Config{
		AppName:      "AtomQuest Goal Portal v1.0",
		ServerHeader: "Go-Fiber Engine",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Catch-all uniform JSON error format handler
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allows any frontend to connect
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// 3. Mount standard global architectural middleware
	app.Use(recover.New()) // Recovers from runtime panics without crashing the entire system
	app.Use(logger.New())  // High-visibility logging format for debugging routes

	// 4. Base Health-Check Route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "Engine is running flawlessly",
		})
	})

	// 5. Mount modular API routes
	api.SetupRoutes(app, database)

	// Determine the port context dynamically
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server initializing on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Critical Error: Web server failed to start: %v", err)
	}
}
