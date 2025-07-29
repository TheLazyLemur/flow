package navigation

import "errors"

// Navigation-specific errors
var (
	ErrInvalidPosition  = errors.New("position is outside grid bounds")
	ErrInvalidCost      = errors.New("cost must be non-negative")
	ErrNoPath           = errors.New("no path exists to goal")
	ErrInvalidGoal      = errors.New("goal position is invalid or blocked")
	ErrEmptyGrid        = errors.New("grid is empty or not initialized")
	ErrInvalidDirection = errors.New("invalid direction")
)