package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"flow/navigation"
)

type BuildingSystem struct {
	turretSystem *TurretSystem
	navigator    *navigation.FlowFieldNavigator
	config       Config
}

func NewBuildingSystem(nav *navigation.FlowFieldNavigator, turretSys *TurretSystem, cfg Config) *BuildingSystem {
	return &BuildingSystem{
		turretSystem: turretSys,
		navigator:    nav,
		config:       cfg,
	}
}

func (bs *BuildingSystem) PlaceBuilding(gridX, gridY int) bool {
	pos := navigation.Position{X: gridX, Y: gridY}
	
	if !bs.navigator.GetGrid().IsValidPosition(pos) {
		return false
	}
	
	if !bs.navigator.GetGrid().IsPassable(pos) {
		return false
	}
	
	// Check if turret already exists at this position
	for _, turret := range bs.turretSystem.Turrets {
		if turret.PositionX == gridX && turret.PositionY == gridY {
			return false
		}
	}
	
	// Create turret
	turret := Turret{
		PositionX:   gridX,
		PositionY:   gridY,
		AttackRange: 3,
		AttackSpeed: 1.0,
	}
	bs.turretSystem.Turrets = append(bs.turretSystem.Turrets, turret)
	
	bs.updateNavigationCosts()
	
	return true
}

func (bs *BuildingSystem) updateNavigationCosts() {
	grid := bs.navigator.GetGrid()
	costs := make([][]int, grid.Height)
	
	for y := range grid.Height {
		costs[y] = make([]int, grid.Width)
		copy(costs[y], grid.Costs[y])
	}
	
	for _, turret := range bs.turretSystem.Turrets {
		if turret.PositionX < grid.Width && turret.PositionY < grid.Height {
			costs[turret.PositionY][turret.PositionX] = -1
			grid.SetBuilding(navigation.Position{X: turret.PositionX, Y: turret.PositionY})
		}
	}
	
	bs.navigator.UpdateCosts(costs)
}

func (bs *BuildingSystem) Draw() {
	for _, turret := range bs.turretSystem.Turrets {
		cellX := int32(bs.config.MarginX + turret.PositionX*bs.config.CellSize)
		cellY := int32(bs.config.MarginY + turret.PositionY*bs.config.CellSize)
		
		rl.DrawRectangle(cellX, cellY, int32(bs.config.CellSize), int32(bs.config.CellSize), rl.Blue)
		
		rl.DrawRectangleLines(cellX, cellY, int32(bs.config.CellSize), int32(bs.config.CellSize), rl.DarkBlue)
	}
}

func (bs *BuildingSystem) GetTurretSystem() *TurretSystem {
	return bs.turretSystem
}