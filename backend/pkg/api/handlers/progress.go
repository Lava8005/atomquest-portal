/*
Progress Handlers for Phase 2 Check-ins.
Humanized top-level documentation: Utilizes PostgreSQL UPSERTs (Insert on Conflict) to handle multiple employee edits during an active quarter window without duplicating data.
*/

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ProgressUpdateRequest struct {
	GoalID            int     `json:"goal_id"`
	Quarter           string  `json:"quarter"` // Expects "Q1", "Q2", "Q3", "Q4"
	ActualAchievement float64 `json:"actual_achievement"`
	Status            string  `json:"status"` // "Not Started", "On Track", "Completed"
}

type ManagerReviewRequest struct {
	SheetID  int    `json:"sheet_id"`
	Quarter  string `json:"quarter"`
	Comments string `json:"comments"`
}

// UpdateProgress handles Employees logging their quarterly actuals
func UpdateProgress(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// In a full implementation, you'd verify the quarter window is currently open here [cite: 36, 37]

		var req ProgressUpdateRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// Perform an UPSERT: Insert the new progress record, or update if that quarter already exists
		_, err := db.Exec(`
			INSERT INTO goal_progress (goal_id, quarter, actual_achievement, status, updated_at)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
			ON CONFLICT (goal_id, quarter) 
			DO UPDATE SET actual_achievement = EXCLUDED.actual_achievement, 
			              status = EXCLUDED.status, 
			              updated_at = CURRENT_TIMESTAMP
		`, req.GoalID, req.Quarter, req.ActualAchievement, req.Status)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log progress"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Quarterly progress logged successfully",
		})
	}
}

// AddManagerReview handles Managers adding structured discussion comments [cite: 32]
func AddManagerReview(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reviewerID := c.Locals("user_id").(int)

		var req ManagerReviewRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
		}

		// UPSERT for manager comments
		_, err := db.Exec(`
			INSERT INTO manager_reviews (sheet_id, quarter, reviewer_id, comments, created_at)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
			ON CONFLICT (sheet_id, quarter)
			DO UPDATE SET comments = EXCLUDED.comments, 
			              reviewer_id = EXCLUDED.reviewer_id,
			              created_at = CURRENT_TIMESTAMP
		`, req.SheetID, req.Quarter, reviewerID, req.Comments)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save check-in comments"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Manager check-in comments securely logged",
		})
	}
}
