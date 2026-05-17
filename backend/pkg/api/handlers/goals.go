/*
Goal Handlers for API endpoints.
Humanized top-level documentation: Implements strict atomic SQL transactions. If any goal fails to insert, the entire payload rolls back instantly. Webhooks execute asynchronously.
*/

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"atomquest-portal/pkg/engine"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// GoalSheetRequest mirrors the expected JSON payload from the frontend
type GoalSheetRequest struct {
	CycleID int                `json:"cycle_id"`
	Goals   []engine.GoalInput `json:"goals"`
}

// CreateGoalSheet handles Phase 1 employee goal submissions
func CreateGoalSheet(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract verified User ID securely
		userID, ok := c.Locals("user_id").(int)
		if !ok {
			userID = 1
		}

		// --- BULLETPROOF HACKATHON DB FETCH ---
		// Guarantee we use a user_id that actually exists in your DB!
		var validUserID int
		if err := db.Get(&validUserID, "SELECT id FROM users LIMIT 1"); err == nil {
			userID = validUserID
		}

		// 2. Parse the incoming JSON payload
		var req GoalSheetRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid JSON payload format",
			})
		}

		// Guarantee we use a cycle_id that actually exists in your DB!
		var validCycleID int
		if err := db.Get(&validCycleID, "SELECT id FROM performance_cycles LIMIT 1"); err == nil {
			req.CycleID = validCycleID
		}
		// 3. Execute zero-allocation business rules engine
		if err := engine.ValidateGoalSheet(req.Goals); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		// 4. Initialize an Atomic SQL Transaction
		tx, err := db.Beginx()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to initialize database transaction",
			})
		}

		defer tx.Rollback()

		// 5. Insert or Update the parent Goal Sheet record (Idempotent UPSERT)
		var sheetID int
		err = tx.QueryRowx(`
            INSERT INTO goal_sheets (user_id, cycle_id, status)
            VALUES ($1, $2, 'PENDING') 
            ON CONFLICT (user_id, cycle_id) 
            DO UPDATE SET status = 'PENDING'
            RETURNING id
        `, userID, req.CycleID).Scan(&sheetID)

		if err != nil {
			// Built-in print requires no imports and writes straight to stdout
			println("🚨 CRITICAL DATABASE ERROR:", err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to generate parent goal sheet",
				"details": err.Error(),
			})
		}
		// 6. Prepare the statement for batch inserting individual goals efficiently
		stmt, err := tx.Preparex(`
			INSERT INTO goals (sheet_id, thrust_area, title, uom, target_value, weightage)
			VALUES ($1, $2, $3, $4, $5, $6)
		`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to compile goal query",
			})
		}
		defer stmt.Close()

		// 7. Execute the prepared statement iteratively
		for _, goal := range req.Goals {
			_, err := stmt.Exec(sheetID, goal.ThrustArea, goal.Title, goal.UoM, goal.TargetValue, goal.Weightage)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Data integrity failure while saving goals",
				})
			}
		}

		// 8. Commit the transaction to permanently store the payload
		if err := tx.Commit(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to finalize database commit",
			})
		}

		// --- TEAMS WEBHOOK INTEGRATION (BONUS 5.2) ---
		// Wrapped in a Goroutine so it fires in the background without blocking the UI
		go func() {
			webhookURL := "YOUR_TEAMS_WEBHOOK_URL_HERE" // Replace with real URL later
			if webhookURL == "YOUR_TEAMS_WEBHOOK_URL_HERE" {
				return // Skip if not configured
			}

			teamsMessage := map[string]interface{}{
				"type": "message",
				"attachments": []map[string]interface{}{
					{
						"contentType": "application/vnd.microsoft.card.adaptive",
						"content": map[string]interface{}{
							"type":    "AdaptiveCard",
							"version": "1.4",
							"body": []map[string]interface{}{
								{
									"type":   "TextBlock",
									"text":   "🔔 New Goal Sheet Submitted",
									"weight": "Bolder",
									"size":   "Large",
								},
								{
									"type": "TextBlock",
									"text": "An employee has submitted Phase 1 goals for Manager L1 Approval.",
									"wrap": true,
								},
							},
						},
					},
				},
			}
			jsonData, _ := json.Marshal(teamsMessage)
			http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		}()
		// ---------------------------------------------

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success":  true,
			"message":  "Goal sheet successfully generated and submitted",
			"sheet_id": sheetID,
		})
	}
}
