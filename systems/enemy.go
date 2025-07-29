package systems

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"flow/navigation"
)

// Enemy represents an animated agent that follows the flow field
type Enemy struct {
	Position  rl.Vector2 // Current position in pixels
	Velocity  rl.Vector2 // Current velocity for smooth movement
	GridPos   rl.Vector2 // Current grid cell position (as floats for easier conversion)
	TargetPos rl.Vector2 // Target position for smooth movement
	Moving    bool       // Whether the unit is currently moving
	Radius    float32    // Unit collision radius
}

// EnemySystem manages all enemy units and their behaviors
type EnemySystem struct {
	enemies   []*Enemy
	navigator *navigation.FlowFieldNavigator
	config    Config
}

// Config holds the configuration for enemy behaviors
type Config struct {
	// Grid dimensions
	Width, Height int
	CellSize      int
	MarginX       int
	MarginY       int

	// Movement parameters
	UnitSpeed float32

	// Steering behavior parameters
	SeparationRadius float32
	SeparationForce  float32
	AlignmentRadius  float32
	AlignmentForce   float32
	CohesionRadius   float32
	CohesionForce    float32
	MaxSteerForce    float32
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		Width:    10,
		Height:   10,
		CellSize: 50,
		MarginX:  30,
		MarginY:  30,

		UnitSpeed: 2.0,

		SeparationRadius: 15.0,
		SeparationForce:  2.0,
		AlignmentRadius:  25.0,
		AlignmentForce:   0.3,
		CohesionRadius:   35.0,
		CohesionForce:    0.2,
		MaxSteerForce:    0.8,
	}
}

// NewEnemySystem creates a new enemy management system
func NewEnemySystem(navigator *navigation.FlowFieldNavigator, config Config) *EnemySystem {
	return &EnemySystem{
		enemies:   make([]*Enemy, 0),
		navigator: navigator,
		config:    config,
	}
}

// SpawnEnemies creates the specified number of enemies
func (es *EnemySystem) SpawnEnemies(count int) {
	for range count {
		// Spread units across the bottom area
		startX := float32(rl.GetRandomValue(0, int32(es.config.Width-1)))
		startY := float32(rl.GetRandomValue(int32(es.config.Height-3), int32(es.config.Height-1)))

		enemy := &Enemy{
			GridPos:  rl.Vector2{X: startX, Y: startY},
			Velocity: rl.Vector2{X: 0, Y: 0},
			Moving:   false,
			Radius:   4.0,
		}

		// Set initial pixel position with small random offset
		enemy.Position = rl.Vector2{
			X: float32(
				es.config.MarginX,
			) + startX*float32(
				es.config.CellSize,
			) + float32(
				es.config.CellSize,
			)/2 + float32(
				rl.GetRandomValue(-10, 10),
			),
			Y: float32(
				es.config.MarginY,
			) + startY*float32(
				es.config.CellSize,
			) + float32(
				es.config.CellSize,
			)/2 + float32(
				rl.GetRandomValue(-10, 10),
			),
		}
		enemy.TargetPos = enemy.Position

		es.enemies = append(es.enemies, enemy)
	}
}

// Update updates all enemies with steering behaviors
func (es *EnemySystem) Update() {
	for _, enemy := range es.enemies {
		// Calculate steering forces
		separation := es.calculateSeparation(enemy)
		alignment := es.calculateAlignment(enemy)
		cohesion := es.calculateCohesion(enemy)
		obstacleAvoid := es.calculateObstacleAvoidance(enemy)

		// Get flow field direction
		flowForce := es.calculateFlowForce(enemy)

		// Combine all forces (flow field has MUCH higher weight for pathfinding)
		totalForce := rl.Vector2{
			X: flowForce.X*5.0 + separation.X*0.5 + alignment.X*0.2 + cohesion.X*0.1 + obstacleAvoid.X*10.0,
			Y: flowForce.Y*5.0 + separation.Y*0.5 + alignment.Y*0.2 + cohesion.Y*0.1 + obstacleAvoid.Y*10.0,
		}

		// Apply force to velocity
		enemy.Velocity.X += totalForce.X * es.config.MaxSteerForce
		enemy.Velocity.Y += totalForce.Y * es.config.MaxSteerForce

		// Limit velocity to max speed
		speed := rl.Vector2Length(enemy.Velocity)
		if speed > es.config.UnitSpeed {
			enemy.Velocity.X = (enemy.Velocity.X / speed) * es.config.UnitSpeed
			enemy.Velocity.Y = (enemy.Velocity.Y / speed) * es.config.UnitSpeed
		}

		// Update position
		enemy.Position.X += enemy.Velocity.X
		enemy.Position.Y += enemy.Velocity.Y

		// Update grid position
		enemy.GridPos.X = (enemy.Position.X - float32(es.config.MarginX) - float32(es.config.CellSize)/2) / float32(
			es.config.CellSize,
		)
		enemy.GridPos.Y = (enemy.Position.Y - float32(es.config.MarginY) - float32(es.config.CellSize)/2) / float32(
			es.config.CellSize,
		)

		// Check if reached goal and reset
		currentGoal := es.navigator.GetGoal()
		if int(enemy.GridPos.X) == currentGoal.X && int(enemy.GridPos.Y) == currentGoal.Y {
			// Reset to random bottom position
			startX := float32(rl.GetRandomValue(0, int32(es.config.Width-1)))
			startY := float32(
				rl.GetRandomValue(int32(es.config.Height-3), int32(es.config.Height-1)),
			)
			enemy.GridPos = rl.Vector2{X: startX, Y: startY}
			enemy.Position = rl.Vector2{
				X: float32(
					es.config.MarginX,
				) + startX*float32(
					es.config.CellSize,
				) + float32(
					es.config.CellSize,
				)/2,
				Y: float32(
					es.config.MarginY,
				) + startY*float32(
					es.config.CellSize,
				) + float32(
					es.config.CellSize,
				)/2,
			}
			enemy.Velocity = rl.Vector2{X: 0, Y: 0}
		}
	}
}

// Draw renders all enemies
func (es *EnemySystem) Draw() {
	for _, enemy := range es.enemies {
		// Draw enemy as a red circle with black outline
		rl.DrawCircle(int32(enemy.Position.X), int32(enemy.Position.Y), enemy.Radius, rl.Red)
		rl.DrawCircleLines(int32(enemy.Position.X), int32(enemy.Position.Y), enemy.Radius, rl.Black)

		// Draw velocity direction line
		if rl.Vector2Length(enemy.Velocity) > 0.1 {
			endX := enemy.Position.X + enemy.Velocity.X*5
			endY := enemy.Position.Y + enemy.Velocity.Y*5
			rl.DrawLine(
				int32(enemy.Position.X),
				int32(enemy.Position.Y),
				int32(endX),
				int32(endY),
				rl.Black,
			)
		}
	}
}

// GetEnemies returns all enemies (for external systems that might need access)
func (es *EnemySystem) GetEnemies() []*Enemy {
	return es.enemies
}

// calculateSeparation keeps enemies from overlapping
func (es *EnemySystem) calculateSeparation(enemy *Enemy) rl.Vector2 {
	steer := rl.Vector2{X: 0, Y: 0}
	count := 0

	// Only check nearby enemies for performance
	for _, other := range es.enemies {
		if other == enemy {
			continue
		}

		// Quick distance check to avoid expensive calculations
		dx := enemy.Position.X - other.Position.X
		dy := enemy.Position.Y - other.Position.Y
		if abs(dx) > es.config.SeparationRadius || abs(dy) > es.config.SeparationRadius {
			continue
		}

		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist > 0 && dist < es.config.SeparationRadius {
			// Calculate repulsion force
			// Normalize and weight by distance (closer = stronger)
			force := es.config.SeparationRadius - dist
			dx = (dx / dist) * force / es.config.SeparationRadius
			dy = (dy / dist) * force / es.config.SeparationRadius
			steer.X += dx
			steer.Y += dy
			count++
		}
	}

	if count > 0 {
		steer.X *= es.config.SeparationForce
		steer.Y *= es.config.SeparationForce
	}

	return steer
}

// calculateAlignment aligns enemy velocity with nearby enemies
func (es *EnemySystem) calculateAlignment(enemy *Enemy) rl.Vector2 {
	steer := rl.Vector2{X: 0, Y: 0}
	count := 0

	for _, other := range es.enemies {
		if other == enemy {
			continue
		}

		dist := rl.Vector2Distance(enemy.Position, other.Position)
		if dist > 0 && dist < es.config.AlignmentRadius {
			steer.X += other.Velocity.X
			steer.Y += other.Velocity.Y
			count++
		}
	}

	if count > 0 {
		steer.X = (steer.X/float32(count) - enemy.Velocity.X) * es.config.AlignmentForce
		steer.Y = (steer.Y/float32(count) - enemy.Velocity.Y) * es.config.AlignmentForce
	}

	return steer
}

// calculateCohesion pulls enemy toward center of nearby enemies
func (es *EnemySystem) calculateCohesion(enemy *Enemy) rl.Vector2 {
	center := rl.Vector2{X: 0, Y: 0}
	count := 0

	for _, other := range es.enemies {
		if other == enemy {
			continue
		}

		dist := rl.Vector2Distance(enemy.Position, other.Position)
		if dist > 0 && dist < es.config.CohesionRadius {
			center.X += other.Position.X
			center.Y += other.Position.Y
			count++
		}
	}

	steer := rl.Vector2{X: 0, Y: 0}
	if count > 0 {
		center.X /= float32(count)
		center.Y /= float32(count)

		// Steer toward center
		steer.X = (center.X - enemy.Position.X) * es.config.CohesionForce
		steer.Y = (center.Y - enemy.Position.Y) * es.config.CohesionForce
	}

	return steer
}

// calculateObstacleAvoidance keeps enemies away from walls
func (es *EnemySystem) calculateObstacleAvoidance(enemy *Enemy) rl.Vector2 {
	steer := rl.Vector2{X: 0, Y: 0}
	grid := es.navigator.GetGrid()

	// Check cells around the enemy
	checkRadius := float32(1.5)
	gridX := int(enemy.GridPos.X)
	gridY := int(enemy.GridPos.Y)

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			checkX := gridX + dx
			checkY := gridY + dy

			// Check if this cell is an obstacle
			if checkX >= 0 && checkX < es.config.Width && checkY >= 0 && checkY < es.config.Height {
				if grid.Costs[checkY][checkX] == -1 {
					// Calculate repulsion from obstacle
					obstacleX := float32(
						es.config.MarginX + checkX*es.config.CellSize + es.config.CellSize/2,
					)
					obstacleY := float32(
						es.config.MarginY + checkY*es.config.CellSize + es.config.CellSize/2,
					)

					dx := enemy.Position.X - obstacleX
					dy := enemy.Position.Y - obstacleY
					dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

					if dist < float32(es.config.CellSize)*checkRadius {
						// Stronger force when closer
						force := (float32(es.config.CellSize)*checkRadius - dist) / (float32(es.config.CellSize) * checkRadius)
						steer.X += (dx / dist) * force
						steer.Y += (dy / dist) * force
					}
				}
			}
		}
	}

	return steer
}

// calculateFlowForce gets the flow field direction for the enemy
func (es *EnemySystem) calculateFlowForce(enemy *Enemy) rl.Vector2 {
	// Get current grid position
	gridX := int(enemy.GridPos.X)
	gridY := int(enemy.GridPos.Y)

	// Bounds check
	if gridX < 0 || gridX >= es.config.Width || gridY < 0 || gridY >= es.config.Height {
		return rl.Vector2{X: 0, Y: 0}
	}

	currentPos := navigation.Position{X: gridX, Y: gridY}
	flowDir, err := es.navigator.GetFlowDirection(currentPos)
	if err != nil {
		return rl.Vector2{X: 0, Y: 0}
	}

	// Convert grid direction to smooth force with proper strength
	return rl.Vector2{
		X: float32(flowDir.X) * 0.8,
		Y: float32(flowDir.Y) * 0.8,
	}
}

// abs returns absolute value of float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

