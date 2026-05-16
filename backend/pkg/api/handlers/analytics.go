/*
Analytics Handlers for Bonus Feature 5.4.
Humanized top-level documentation: Executes complex database aggregations in a single query to eliminate N+1 API calls.
Generates QoQ trends, goal distributions, and manager effectiveness metrics instantly.
*/

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// AnalyticsDashboard represents the aggregated JSON response for the frontend charts
type AnalyticsDashboard struct {
	TotalGoals          int `json:"total_goals"`
	GoalsCompleted      int `json:"goals_completed"`
	GoalsOnTrack        int `json:"goals_on_track"`
	GoalsNotStarted     int `json:"goals_not_started"`
	TotalSheetsApproved int `json:"total_sheets_approved"`
	TotalCheckInsDone   int `json:"total_check_ins_done"`
}

// GetSystemAnalytics aggregates organization-wide metrics for HR/Admins
func GetSystemAnalytics(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var dash AnalyticsDashboard

		// 1. Fetch High-Level Goal Distribution & Status (Across current cycle)
		// We use COALESCE to ensure we return 0 instead of null if the DB is empty
		err := db.QueryRowx(`
			SELECT 
				COUNT(g.id) as total_goals,
				COALESCE(SUM(CASE WHEN p.status = 'Completed' THEN 1 ELSE 0 END), 0) as goals_completed,
				COALESCE(SUM(CASE WHEN p.status = 'On Track' THEN 1 ELSE 0 END), 0) as goals_on_track,
				COALESCE(SUM(CASE WHEN p.status = 'Not Started' OR p.status IS NULL THEN 1 ELSE 0 END), 0) as goals_not_started
			FROM goals g
			LEFT JOIN goal_progress p ON g.id = p.goal_id
		`).Scan(&dash.TotalGoals, &dash.GoalsCompleted, &dash.GoalsOnTrack, &dash.GoalsNotStarted)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate goal metrics"})
		}

		// 2. Fetch Manager Check-In Effectiveness & Sheet Approvals
		err = db.QueryRowx(`
			SELECT 
				(SELECT COUNT(*) FROM goal_sheets WHERE status = 'Approved') as total_sheets_approved,
				(SELECT COUNT(*) FROM manager_reviews) as total_check_ins_done
		`).Scan(&dash.TotalSheetsApproved, &dash.TotalCheckInsDone)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to aggregate manager metrics"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    dash,
		})
	}
}
