/*
Engine package provides standalone, low-overhead business logic evaluation.
Humanized top-level documentation: Zero intermediate state creation to maximize O(1) space complexity.
*/

package engine

import (
	"errors"
	"time"
)

// Structural representations mapping cleanly to our PostgreSQL data types
type GoalUoM string
type GoalStatus string

const (
	Numeric    GoalUoM = "Numeric"
	Percentage GoalUoM = "Percentage"
	Timeline   GoalUoM = "Timeline"
	ZeroBased  GoalUoM = "Zero-based"

	NotStarted GoalStatus = "Not Started"
	OnTrack    GoalStatus = "On Track"
	Completed  GoalStatus = "Completed"
)

type GoalInput struct {
	ThrustArea  string  `json:"thrust_area"`
	Title       string  `json:"title"`
	UoM         GoalUoM `json:"uom"`
	TargetValue float64 `json:"target_value"`
	Weightage   int     `json:"weightage"`
}

type ProgressMetrics struct {
	UoM              GoalUoM
	TargetValue      float64
	ActualValue      float64
	Deadline         time.Time
	ActualCompletion time.Time
}

var (
	ErrMaxGoalsExceeded  = errors.New("maximum number of goals per employee is 8")
	ErrMinWeightInvalid  = errors.New("minimum weightage per individual goal must be 10%")
	ErrTotalWeightMismat = errors.New("total weightage across all goals must equal 100%")
)

// ValidateGoalSheet checks if the compiled goal array conforms to strict business rules
func ValidateGoalSheet(goals []GoalInput) error {
	// 1. System-enforced capacity check (Max 8 goals)
	if len(goals) > 8 {
		return ErrMaxGoalsExceeded
	}

	totalWeight := 0
	for _, goal := range goals {
		// 2. System-enforced individual floor validation (Min 10%)
		if goal.Weightage < 10 {
			return ErrMinWeightInvalid
		}
		totalWeight += goal.Weightage
	}

	// 3. System-enforced cumulative structural integrity validation (Exactly 100%)
	if totalWeight != 100 {
		return ErrTotalWeightMismat
	}

	return nil
}

// ComputeProgressScore processes tracking metrics mathematically according to UoM types
func ComputeProgressScore(m ProgressMetrics, isLowerBetter bool) float64 {
	switch m.UoM {
	case Numeric, Percentage:
		if m.TargetValue == 0 {
			return 0.0
		}
		if isLowerBetter {
			// Max configuration: Lower is better (e.g., TAT, Cost) -> Target ÷ Achievement
			return (m.TargetValue / m.ActualValue) * 100.0
		}
		// Min configuration: Higher is better (e.g., Sales Revenue) -> Achievement ÷ Target
		return (m.ActualValue / m.TargetValue) * 100.0

	case Timeline:
		// Date-based completion comparison
		if m.ActualCompletion.Before(m.Deadline) || m.ActualCompletion.Equal(m.Deadline) {
			return 100.0
		}
		return 0.0

	case ZeroBased:
		// Zero configuration: Zero equals absolute success (e.g., Safety incidents)
		if m.ActualValue == 0 {
			return 100.0
		}
		return 0.0
	}

	return 0.0
}
