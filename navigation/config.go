package navigation

import "errors"

// Config contains configuration for the navigation system
type Config struct {
	// Grid dimensions
	GridWidth  int
	GridHeight int

	// Movement directions (4-way or 8-way)
	Directions []Direction

	// Cost multiplier for diagonal movements (typically sqrt(2) ≈ 1.4)
	DiagonalCost float64

	// Whether to allow diagonal movement through corners
	AllowCornerCutting bool
}

// EightWayConfig returns a configuration for 8-way movement
func EightWayConfig(width, height int) Config {
	return Config{
		GridWidth:          width,
		GridHeight:         height,
		Directions:         EightWayDirections,
		DiagonalCost:       1.4,
		AllowCornerCutting: true,
	}
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.GridWidth <= 0 || c.GridHeight <= 0 {
		return errors.New("grid dimensions must be positive")
	}

	if len(c.Directions) == 0 {
		return errors.New("must have at least one direction")
	}

	if c.DiagonalCost <= 0 {
		return errors.New("diagonal cost must be positive")
	}

	return nil
}

