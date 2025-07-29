// This program implements a flow field pathfinding algorithm with raylib visualization.
// Flow fields are efficient for multiple agents navigating to the same goal,
// as the pathfinding calculation is done once for the entire grid.
// Each cell stores a direction vector pointing toward the optimal path to the goal.
package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"

	"flow/navigation"
	"flow/systems"
)

// Graphics constants for raylib visualization
const (
	cellSize = 50 // Size of each grid cell in pixels
	marginX  = 30 // Left/right margin
	marginY  = 30 // Top/bottom margin
	fontSize = 24 // Font size for arrows
)

var (
	// Grid dimensions
	Width, Height = 10, 10

	// Window dimensions calculated from grid size
	windowWidth  = Width*cellSize + 2*marginX
	windowHeight = Height*cellSize + 2*marginY

	// Navigation system
	navigator *navigation.FlowFieldNavigator
	// Enemy system
	enemySystem *systems.EnemySystem
)

func main() {
	// Initialize navigation system
	config := navigation.EightWayConfig(Width, Height)
	var err error
	navigator, err = navigation.NewFlowFieldNavigator(config)
	if err != nil {
		log.Fatal("Failed to create navigator:", err)
	}

	// Set up initial obstacles
	setupObstacles()

	// Set initial goal
	initialGoal := navigation.Position{X: 7, Y: 2}
	if err := navigator.SetGoal(initialGoal); err != nil {
		log.Fatal("Failed to set initial goal:", err)
	}

	// Initialize raylib window for graphics visualization
	rl.InitWindow(int32(windowWidth), int32(windowHeight), "Flow Field Pathfinding Visualization")
	defer rl.CloseWindow()

	// Set target FPS for smooth rendering
	rl.SetTargetFPS(60)

	// Initialize enemy system
	enemyConfig := systems.Config{
		Width:            Width,
		Height:           Height,
		CellSize:         cellSize,
		MarginX:          marginX,
		MarginY:          marginY,
		UnitSpeed:        2.0,
		SeparationRadius: 15.0,
		SeparationForce:  10.0,
		AlignmentRadius:  25.0,
		AlignmentForce:   0.3,
		CohesionRadius:   35.0,
		CohesionForce:    0.2,
		MaxSteerForce:    0.8,
	}
	enemySystem = systems.NewEnemySystem(navigator, enemyConfig)
	enemySystem.SpawnEnemies(100)

	// Main rendering loop
	for !rl.WindowShouldClose() {
		// Handle mouse input for goal placement
		handleMouseInput()

		// Update all enemies with steering behaviors
		enemySystem.Update()

		// Begin drawing phase
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw the flow field grid
		drawFlowField()

		// Draw all enemies
		enemySystem.Draw()

		// End drawing phase
		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}
}

// setupObstacles creates initial obstacles in the grid
func setupObstacles() {
	grid := navigator.GetGrid()
	costs := make([][]int, Height)

	// Initialize costs with current grid state
	for y := range Height {
		costs[y] = make([]int, Width)
		copy(costs[y], grid.Costs[y])
	}

	// Add obstacles: Create a 3x3 wall from (4,4) to (6,6)
	for x := 4; x <= 6; x++ {
		for y := 4; y <= 6; y++ {
			if x < Width && y < Height {
				costs[y][x] = -1 // Mark as obstacle
			}
		}
	}

	// Update navigator with new costs
	if err := navigator.UpdateCosts(costs); err != nil {
		log.Printf("Failed to update costs: %v", err)
	}
}

// handleMouseInput checks for mouse clicks and updates goal position
func handleMouseInput() {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		mousePos := rl.GetMousePosition()

		gridX := int((mousePos.X - float32(marginX)) / float32(cellSize))
		gridY := int((mousePos.Y - float32(marginY)) / float32(cellSize))

		newGoal := navigation.Position{X: gridX, Y: gridY}

		// Try to set the new goal
		if err := navigator.SetGoal(newGoal); err != nil {
			// Goal is invalid (out of bounds or obstacle), ignore click
			return
		}
	}
}

// drawFlowField renders the entire flow field grid using raylib
func drawFlowField() {
	// Draw grid background and cell borders
	drawGrid()

	goal := navigator.GetGoal()
	grid := navigator.GetGrid()

	// Draw each cell based on its type and flow direction
	for y := range Height {
		for x := range Width {
			// Calculate screen position for this grid cell
			cellX := int32(marginX + x*cellSize)
			cellY := int32(marginY + y*cellSize)

			if grid.Costs[y][x] == -1 {
				// Draw obstacles as black filled rectangles
				rl.DrawRectangle(cellX, cellY, int32(cellSize), int32(cellSize), rl.Black)
			} else if x == goal.X && y == goal.Y {
				// Draw goal as bright green rectangle
				rl.DrawRectangle(cellX, cellY, int32(cellSize), int32(cellSize), rl.Lime)
				// Add "GOAL" text in center
				textWidth := rl.MeasureText("GOAL", int32(fontSize/2))
				rl.DrawText("GOAL", cellX+(int32(cellSize)-textWidth)/2, cellY+int32(cellSize)/2-int32(fontSize/4), int32(fontSize/2), rl.Black)
			} else {
				// Draw flow arrow for navigable cells
				direction := grid.FlowField[y][x]
				drawFlowArrow(cellX, cellY, direction)
			}
		}
	}
}

// drawGrid renders the background grid lines for visual clarity
func drawGrid() {
	// Draw vertical grid lines
	for x := 0; x <= Width; x++ {
		lineX := int32(marginX + x*cellSize)
		rl.DrawLine(lineX, int32(marginY), lineX, int32(marginY+Height*cellSize), rl.LightGray)
	}

	// Draw horizontal grid lines
	for y := 0; y <= Height; y++ {
		lineY := int32(marginY + y*cellSize)
		rl.DrawLine(int32(marginX), lineY, int32(marginX+Width*cellSize), lineY, rl.LightGray)
	}
}

// drawFlowArrow renders a directional arrow in the specified cell
func drawFlowArrow(cellX, cellY int32, direction navigation.Direction) {
	// Skip drawing if no direction
	if direction.X == 0 && direction.Y == 0 {
		// Draw a dot for unreachable cells
		centerX := cellX + int32(cellSize)/2
		centerY := cellY + int32(cellSize)/2
		rl.DrawCircle(centerX, centerY, 3, rl.Gray)
		return
	}

	// Calculate center of cell
	centerX := cellX + int32(cellSize)/2
	centerY := cellY + int32(cellSize)/2

	// Arrow dimensions
	arrowLength := int32(cellSize / 3)
	arrowHeadSize := int32(cellSize / 8)

	// Calculate arrow end point based on direction
	endX := centerX + int32(float32(direction.X)*float32(arrowLength))
	endY := centerY + int32(float32(direction.Y)*float32(arrowLength))

	// Draw arrow shaft (line from center towards direction)
	rl.DrawLine(centerX, centerY, endX, endY, rl.DarkBlue)

	// Calculate arrow head points
	// Arrow head is perpendicular to the direction
	perpX := float32(-direction.Y)
	perpY := float32(direction.X)

	// Arrow head triangle points
	head1X := endX - int32(
		float32(direction.X)*float32(arrowHeadSize),
	) + int32(
		perpX*float32(arrowHeadSize)/2,
	)
	head1Y := endY - int32(
		float32(direction.Y)*float32(arrowHeadSize),
	) + int32(
		perpY*float32(arrowHeadSize)/2,
	)
	head2X := endX - int32(
		float32(direction.X)*float32(arrowHeadSize),
	) - int32(
		perpX*float32(arrowHeadSize)/2,
	)
	head2Y := endY - int32(
		float32(direction.Y)*float32(arrowHeadSize),
	) - int32(
		perpY*float32(arrowHeadSize)/2,
	)

	// Draw arrow head triangle
	rl.DrawTriangle(
		rl.Vector2{X: float32(endX), Y: float32(endY)},
		rl.Vector2{X: float32(head1X), Y: float32(head1Y)},
		rl.Vector2{X: float32(head2X), Y: float32(head2Y)},
		rl.DarkBlue,
	)
}
