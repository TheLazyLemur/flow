package navigation

// Position represents a grid coordinate position
type Position struct {
	X, Y int
}

// Direction represents a movement direction vector
type Direction struct {
	X, Y int
}

// CellType represents the type of a grid cell
type CellType int

const (
	Passable CellType = iota
	Obstacle
	Goal
	Building
)

// Standard direction sets for different movement patterns
var (
	// FourWayDirections allows only cardinal movement (up, down, left, right)
	FourWayDirections = []Direction{
		{X: 0, Y: -1}, // Up
		{X: 0, Y: 1},  // Down
		{X: -1, Y: 0}, // Left
		{X: 1, Y: 0},  // Right
	}

	// EightWayDirections allows both cardinal and diagonal movement
	EightWayDirections = []Direction{
		{X: 0, Y: -1},  // Up
		{X: 0, Y: 1},   // Down
		{X: -1, Y: 0},  // Left
		{X: 1, Y: 0},   // Right
		{X: -1, Y: -1}, // Up-Left
		{X: -1, Y: 1},  // Down-Left
		{X: 1, Y: -1},  // Up-Right
		{X: 1, Y: 1},   // Down-Right
	}
)

// Grid represents the navigation grid with costs
type Grid struct {
	Width, Height int
	Costs         [][]int  // -1 for obstacles, positive values for movement cost
	FlowField     [][]Direction
	Distances     [][]int
	CellTypes     [][]CellType
}

// NewGrid creates a new navigation grid with the specified dimensions
func NewGrid(width, height int) *Grid {
	grid := &Grid{
		Width:     width,
		Height:    height,
		Costs:     make([][]int, height),
		FlowField: make([][]Direction, height),
		Distances: make([][]int, height),
		CellTypes: make([][]CellType, height),
	}

	// Initialize all slices
	for y := 0; y < height; y++ {
		grid.Costs[y] = make([]int, width)
		grid.FlowField[y] = make([]Direction, width)
		grid.Distances[y] = make([]int, width)
		grid.CellTypes[y] = make([]CellType, width)
		
		// Initialize with passable terrain (cost = 1)
		for x := 0; x < width; x++ {
			grid.Costs[y][x] = 1
			grid.CellTypes[y][x] = Passable
		}
	}

	return grid
}

// IsValidPosition checks if a position is within grid bounds
func (g *Grid) IsValidPosition(pos Position) bool {
	return pos.X >= 0 && pos.X < g.Width && pos.Y >= 0 && pos.Y < g.Height
}

// IsPassable checks if a position is passable (not an obstacle)
func (g *Grid) IsPassable(pos Position) bool {
	if !g.IsValidPosition(pos) {
		return false
	}
	return g.Costs[pos.Y][pos.X] != -1
}

// SetObstacle marks a position as an obstacle
func (g *Grid) SetObstacle(pos Position) error {
	if !g.IsValidPosition(pos) {
		return ErrInvalidPosition
	}
	g.Costs[pos.Y][pos.X] = -1
	g.CellTypes[pos.Y][pos.X] = Obstacle
	return nil
}

// SetBuilding marks a position as a building
func (g *Grid) SetBuilding(pos Position) error {
	if !g.IsValidPosition(pos) {
		return ErrInvalidPosition
	}
	g.Costs[pos.Y][pos.X] = -1
	g.CellTypes[pos.Y][pos.X] = Building
	return nil
}

// SetCost sets the movement cost for a position
func (g *Grid) SetCost(pos Position, cost int) error {
	if !g.IsValidPosition(pos) {
		return ErrInvalidPosition
	}
	if cost < 0 {
		return ErrInvalidCost
	}
	g.Costs[pos.Y][pos.X] = cost
	return nil
}

// GetFlowDirection returns the flow direction at a given position
func (g *Grid) GetFlowDirection(pos Position) (Direction, error) {
	if !g.IsValidPosition(pos) {
		return Direction{}, ErrInvalidPosition
	}
	return g.FlowField[pos.Y][pos.X], nil
}