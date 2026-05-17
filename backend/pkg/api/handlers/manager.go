package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// PendingSheet defines the JSON structure sent to the frontend
type PendingSheet struct {
	SheetID      int    `json:"id" db:"sheet_id"`
	EmployeeName string `json:"employee_name" db:"employee_name"`
	Department   string `json:"department" db:"department"`
	GoalCount    int    `json:"goal_count" db:"goal_count"`
	TotalWeight  int    `json:"total_weight" db:"total_weight"`
	Status       string `json:"status" db:"status"`
}

// GetPendingGoalSheets queries PostgreSQL for sheets awaiting approval
func GetPendingGoalSheets(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Real SQL query joining Users and their Goal Sheets
		// Adjust table names (goal_sheets, users) if your schema.sql named them differently
		query := `
			SELECT 
				gs.id AS sheet_id,
				u.name AS employee_name,
				u.department,
				COUNT(g.id) AS goal_count,
				SUM(g.weightage) AS total_weight,
				gs.status
			FROM goal_sheets gs
			JOIN users u ON gs.user_id = u.id
			LEFT JOIN goals g ON g.sheet_id = gs.id
			WHERE gs.status = 'PENDING'
			GROUP BY gs.id, u.name, u.department, gs.status
		`

		var pendingSheets []PendingSheet
		err := db.Select(&pendingSheets, query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Database error fetching pending sheets",
				"details": err.Error(),
			})
		}

		// If no sheets are pending, return an empty array instead of null
		if len(pendingSheets) == 0 {
			return c.Status(fiber.StatusOK).JSON([]PendingSheet{})
		}

		return c.Status(fiber.StatusOK).JSON(pendingSheets)
	}
}

// ApproveRequest struct matches the JSON body sent from TypeScript
type ManagerApproveRequest struct {
	Status string `json:"status"`
}

// ApproveGoalSheet executes the UPDATE transaction in PostgreSQL
func ManagerApproveGoalSheet(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sheetID := c.Params("id") // Gets the ID from the URL /sheet/:id/approve

		var req ManagerApproveRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
		}

		// Atomic SQL Transaction to update the status
		query := `UPDATE goal_sheets SET status = $1 WHERE id = $2`
		result, err := db.Exec(query, req.Status, sheetID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to update sheet status",
			})
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Sheet not found"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Goal sheet successfully approved",
		})
	}
}
