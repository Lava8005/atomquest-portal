/*
API Routing Registry.
Humanized top-level documentation: Centralizes all API paths and injects RBAC middleware dynamically.
*/

package api

import (
	"atomquest-portal/pkg/api/handlers"
	"atomquest-portal/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes mounts all grouped endpoints to the Fiber application instance
func SetupRoutes(app *fiber.App, db *sqlx.DB) {
	// Root API Group
	v1 := app.Group("/api/v1")

	// Phase 1: Goal Setting (Protected endpoints)
	goals := v1.Group("/goals", middleware.Protect())

	// Route: POST /api/v1/goals/sheet (Employee submits)
	goals.Post("/sheet", middleware.RequireRoles("Employee"), handlers.CreateGoalSheet(db))

	// Route: GET /api/v1/goals/pending (Manager fetches real DB rows)
	goals.Get("/pending", middleware.RequireRoles("Manager", "Admin"), handlers.GetPendingGoalSheets(db))

	// Route: PUT /api/v1/goals/sheet/:id/approve (Manager updates DB row)
	goals.Put("/sheet/:id/approve", middleware.RequireRoles("Manager", "Admin"), handlers.ManagerApproveGoalSheet(db))
	// Phase 2: Progress Tracking & Check-ins
	progress := v1.Group("/progress", middleware.Protect())

	// Employees update their actuals [cite: 28, 40]
	progress.Post("/update", middleware.RequireRoles("Employee"), handlers.UpdateProgress(db))

	// Managers submit their check-in comments [cite: 32, 40]
	progress.Post("/review", middleware.RequireRoles("Manager", "Admin"), handlers.AddManagerReview(db))
	// Bonus Feature 5.4: Analytics Dashboard
	analytics := v1.Group("/analytics", middleware.Protect())

	// Only Admins can see the global heatmap and QoQ data
	analytics.Get("/dashboard", middleware.RequireRoles("Admin"), handlers.GetSystemAnalytics(db))
	// Bonus Feature 5.3: Rule-Based Escalation Module
	escalations := v1.Group("/escalations", middleware.Protect(), middleware.RequireRoles("Admin"))

	// Manual trigger for the hackathon presentation demo
	escalations.Post("/trigger", handlers.TriggerEscalationEngine(db))

	// Fetch the logs for the HR Dashboard
	escalations.Get("/logs", handlers.GetEscalationLogs(db))
}
