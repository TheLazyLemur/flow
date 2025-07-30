package systems

import (
	"fmt"
	"math"
)

type Turret struct {
	PositionX   int
	PositionY   int
	AttackRange int
	AttackSpeed float64
}

type TurretSystem struct {
	Turrets     []Turret
	enemySystem *EnemySystem
	config      Config
}

func NewTurretSystem(enemySys *EnemySystem, cfg Config) *TurretSystem {
	return &TurretSystem{
		Turrets:     make([]Turret, 0),
		enemySystem: enemySys,
		config:      cfg,
	}
}

func (ts *TurretSystem) Update() {
	for _, turret := range ts.Turrets {
		ts.checkEnemiesInRange(turret)
	}
}

func (ts *TurretSystem) checkEnemiesInRange(turret Turret) {
	enemies := ts.enemySystem.GetEnemies()
	
	for _, enemy := range enemies {
		// Convert enemy screen position to grid position
		enemyGridX := int((enemy.Position.X - float32(ts.config.MarginX)) / float32(ts.config.CellSize))
		enemyGridY := int((enemy.Position.Y - float32(ts.config.MarginY)) / float32(ts.config.CellSize))
		
		// Calculate distance between turret and enemy
		dx := float64(turret.PositionX - enemyGridX)
		dy := float64(turret.PositionY - enemyGridY)
		distance := math.Sqrt(dx*dx + dy*dy)
		
		if distance <= float64(turret.AttackRange) {
			fmt.Println("ENEMY IN RANGE")
		}
	}
}
