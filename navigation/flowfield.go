package navigation

import (
	"errors"
	"math"
)

// FlowFieldNavigator implements pathfinding using flow fields
type FlowFieldNavigator struct {
	config    Config
	grid      *Grid
	goal      Position
	isGoalSet bool
}

// NewFlowFieldNavigator creates a new flow field navigator with the given configuration
func NewFlowFieldNavigator(config Config) (*FlowFieldNavigator, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	grid := NewGrid(config.GridWidth, config.GridHeight)

	return &FlowFieldNavigator{
		config:    config,
		grid:      grid,
		isGoalSet: false,
	}, nil
}

// SetGoal sets the target position and recomputes the flow field
func (f *FlowFieldNavigator) SetGoal(goal Position) error {
	if !f.grid.IsValidPosition(goal) {
		return ErrInvalidPosition
	}

	if !f.grid.IsPassable(goal) {
		return ErrInvalidGoal
	}

	f.goal = goal
	f.isGoalSet = true

	return f.computeFlowField()
}

// GetFlowDirection returns the optimal direction to move from the given position
func (f *FlowFieldNavigator) GetFlowDirection(pos Position) (Direction, error) {
	if !f.isGoalSet {
		return Direction{}, ErrInvalidGoal
	}

	if !f.grid.IsValidPosition(pos) {
		return Direction{}, ErrInvalidPosition
	}

	// If we're at the goal, no movement needed
	if pos.X == f.goal.X && pos.Y == f.goal.Y {
		return Direction{X: 0, Y: 0}, nil
	}

	direction := f.grid.FlowField[pos.Y][pos.X]

	// Check if position is reachable
	if direction.X == 0 && direction.Y == 0 && (pos.X != f.goal.X || pos.Y != f.goal.Y) {
		return Direction{}, ErrNoPath
	}

	return direction, nil
}

// UpdateCosts updates the grid costs and recomputes the flow field if goal is set
func (f *FlowFieldNavigator) UpdateCosts(costs [][]int) error {
	if len(costs) != f.grid.Height {
		return errors.New("cost grid height doesn't match navigator grid")
	}

	for y := range f.grid.Height {
		if len(costs[y]) != f.grid.Width {
			return errors.New("cost grid width doesn't match navigator grid")
		}
		copy(f.grid.Costs[y], costs[y])
	}

	// Recompute flow field if goal is set
	if f.isGoalSet {
		// Check if goal is still valid
		if !f.grid.IsPassable(f.goal) {
			f.isGoalSet = false
			return ErrInvalidGoal
		}

		return f.computeFlowField()
	}

	return nil
}

// GetGoal returns the current goal position
func (f *FlowFieldNavigator) GetGoal() Position {
	return f.goal
}

// GetGrid returns a copy of the current grid state
func (f *FlowFieldNavigator) GetGrid() *Grid {
	// Create a deep copy to prevent external modification
	gridCopy := NewGrid(f.grid.Width, f.grid.Height)

	for y := range f.grid.Height {
		copy(gridCopy.Costs[y], f.grid.Costs[y])
		copy(gridCopy.FlowField[y], f.grid.FlowField[y])
		copy(gridCopy.Distances[y], f.grid.Distances[y])
	}

	return gridCopy
}

// computeFlowField calculates the flow field using Dijkstra's algorithm
func (f *FlowFieldNavigator) computeFlowField() error {
	// Reset distances and flow field
	for y := range f.grid.Height {
		for x := range f.grid.Width {
			f.grid.Distances[y][x] = math.MaxInt32
			f.grid.FlowField[y][x] = Direction{X: 0, Y: 0}
		}
	}

	// Initialize goal
	f.grid.Distances[f.goal.Y][f.goal.X] = 0
	queue := []Position{f.goal}

	// Phase 1: Dijkstra-style distance propagation
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentDist := f.grid.Distances[current.Y][current.X]

		// Check all configured directions
		for _, dir := range f.config.Directions {
			next := Position{
				X: current.X + dir.X,
				Y: current.Y + dir.Y,
			}

			// Skip if out of bounds or blocked
			if !f.grid.IsValidPosition(next) || !f.grid.IsPassable(next) {
				continue
			}

			// Calculate movement cost
			moveCost := f.grid.Costs[next.Y][next.X]

			// Apply diagonal cost multiplier if needed
			if f.isDiagonal(dir) {
				moveCost = int(float64(moveCost) * f.config.DiagonalCost)
			}

			newDist := currentDist + moveCost

			// Update if we found a shorter path
			if newDist < f.grid.Distances[next.Y][next.X] {
				f.grid.Distances[next.Y][next.X] = newDist
				queue = append(queue, next)
			}
		}
	}

	// Phase 2: Compute flow directions
	for y := range f.grid.Height {
		for x := range f.grid.Width {
			pos := Position{X: x, Y: y}

			// Skip obstacles and goal
			if !f.grid.IsPassable(pos) || (x == f.goal.X && y == f.goal.Y) {
				continue
			}

			bestDist := f.grid.Distances[y][x]
			bestDir := Direction{X: 0, Y: 0}

			// Find neighbor with minimum distance
			for _, dir := range f.config.Directions {
				neighbor := Position{X: x + dir.X, Y: y + dir.Y}

				if f.grid.IsValidPosition(neighbor) {
					neighborDist := f.grid.Distances[neighbor.Y][neighbor.X]
					if neighborDist < bestDist {
						bestDist = neighborDist
						bestDir = dir
					}
				}
			}

			f.grid.FlowField[y][x] = bestDir
		}
	}

	return nil
}

// isDiagonal checks if a direction is diagonal
func (f *FlowFieldNavigator) isDiagonal(dir Direction) bool {
	return dir.X != 0 && dir.Y != 0
}
