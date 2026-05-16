/*
Escalation Engine Handlers for Bonus Feature 5.3.
Humanized top-level documentation: Executes bulk threshold checks directly within PostgreSQL to maintain O(1) application memory footprint.
*/

package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type EscalationLog struct {
	ID            int    `json:"id" db:"id"`
	EmployeeName  string `json:"employee_name" db:"employee_name"`
	ManagerName   string `json:"manager_name" db:"manager_name"`
	RuleTriggered string `json:"rule_triggered" db:"rule_triggered"`
	Status        string `json:"status" db:"status"`
	CreatedAt     string `json:"created_at" db:"created_at"`
}

// TriggerEscalationEngine runs the background checks on-demand for easy demoing
func TriggerEscalationEngine(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Rule 1: Manager has not approved goals within 3 days of submission
		// Using ON CONFLICT DO NOTHING ensures we don't spam duplicate logs for the same offense
		_, err := db.Exec(`
			INSERT INTO escalations (user_id, manager_id, rule_triggered)
			SELECT gs.user_id, u.manager_id, 'L1 Approval Overdue (>3 Days)'
			FROM goal_sheets gs
			JOIN users u ON gs.user_id = u.id
			WHERE gs.status = 'Pending Approval' 
			AND gs.updated_at < NOW() - INTERVAL '3 days'
			ON CONFLICT DO NOTHING
		`)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process approval escalations"})
		}

		// Rule 2: Employee has not completed their check-in for an active quarter
		// (Assuming a simple check where they have approved goals but no progress logged for Q1)
		_, err = db.Exec(`
			INSERT INTO escalations (user_id, manager_id, rule_triggered)
			SELECT DISTINCT gs.user_id, u.manager_id, 'Q1 Check-in Overdue'
			FROM goal_sheets gs
			JOIN users u ON gs.user_id = u.id
			LEFT JOIN goal_progress p ON p.goal_id IN (SELECT id FROM goals WHERE sheet_id = gs.id) AND p.quarter = 'Q1'
			WHERE gs.status = 'Approved' AND p.id IS NULL
			ON CONFLICT DO NOTHING
		`)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process check-in escalations"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Escalation rules processed. Violations logged successfully.",
		})
	}
}

// GetEscalationLogs fetches the centralized audit trail for Admin/HR dashboards
func GetEscalationLogs(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var logs []EscalationLog

		// Join with users table twice to resolve both employee and manager names
		err := db.Select(&logs, `
			SELECT 
				e.id, 
				u1.name as employee_name, 
				COALESCE(u2.name, 'Unassigned') as manager_name, 
				e.rule_triggered, 
				e.status, 
				TO_CHAR(e.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at
			FROM escalations e
			JOIN users u1 ON e.user_id = u1.id
			LEFT JOIN users u2 ON e.manager_id = u2.id
			ORDER BY e.created_at DESC
		`)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve escalation logs"})
		}

		// Return an empty array instead of null if no logs exist
		if logs == nil {
			logs = []EscalationLog{}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    logs,
		})
	}
}
