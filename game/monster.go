package game

import (
	"fmt"
	"math"
)

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'R', "Rat"}, "Monster", 5, 1, 1.5, 3, 0.0, true}}
}

func NewSpider(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'S', "Spider"}, "Monster", 7, 0, .25, 5, 0.0, true}}
}

func (m *Monster) Update(level *Level) {
	p := level.Player
	path, _, found := level.astar(m.Pos, p.Pos)
	moveIndex := 1

	if m.Hitpoints < 0 {
		m.Dead(level)
	}

	if found && p.Alive {
		m.AP += m.Speed
		apInt := int(m.AP)
		for i := 0; i < apInt; i++ {
			if moveIndex < len(path) {
				if m.Move(path[moveIndex], level) {
					moveIndex++
				}
			}
		}
	}
}

func (m *Monster) Move(to Pos, level *Level) bool {
	moved := false
	tile := *level.TileAtPos(to)
	if m.Hitpoints > 0 && m.AP >= float64(tile.Cost) {
		_, exists := level.Monsters[to]
		if !exists && to != level.Player.Pos {
			delete(level.Monsters, m.Pos)
			level.Monsters[to] = m
			m.Pos = to
			fmt.Println("moved!")
			m.AP -= float64(tile.Cost)
			moved = true
		} else if to == level.Player.Pos {
			events := Attack(&m.Character, &level.Player.Character)
			level.AddEvents(events...)
		}
	}

	return moved
}

// TODO: consider combining with lineOfSight taking in an character or entity
func (m *Monster) isPlayerInRange(level *Level) bool {
	pos := m.Pos
	dist := m.SightRange
	player := level.Player.Pos

	for y := pos.Y - dist; y <= pos.Y+dist; y++ {
		for x := pos.X - dist; x <= pos.X+dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				line := level.bresenham(pos, Pos{x, y})
				for _, p := range line {
					level.Debug[pos] = true
					if p == player {
						fmt.Println("target found")
						return true
					}
				}
			}
		}
	}
	return false
}

func (m *Monster) Dead(level *Level) {
	level.TileAtPos(m.Pos).BloodStained = true
	level.AddEvents(fmt.Sprintf("%s died", m.Name))
	delete(level.Monsters, m.Pos)
}
