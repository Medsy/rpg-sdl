package game

import (
	"fmt"
	"math"
)

type Monster struct {
	Character
	Path []Pos
}

func NewRat(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'R', "Rat"}, "Monster", 5, 1, 2, 1.5, 3, 0.0, true}, make([]Pos, 0)}
}

func NewSpider(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'S', "Spider"}, "Monster", 7, 0, 5, .5, 5, .5, true}, make([]Pos, 0)} // this is a little fucky wucky
}

func (m *Monster) monsterToString() string {
	return fmt.Sprintf("{name: %s, HP: %d, SightRange: %d, AP: %f, Speed: %f}", m.Name, m.Hitpoints, m.SightRange, m.AP, m.Speed) // TODO: add all monsters stats to be printed
}

func (m *Monster) Update(level *Level) {
	var moveIndex int
	p := level.Player

	if len(m.Path) == 0 {
		m.Path, _, _ = level.astar(m.Pos, p.Pos)
		moveIndex = 1
	}

	if m.Hitpoints < 0 {
		m.Dead(level)
	}

	if len(m.Path) != 0 && p.Alive {
		m.AP += m.Speed
		apInt := int(m.AP)
		for i := 0; i < apInt; i++ {
			if moveIndex < len(m.Path) {
				if m.Move(m.Path[moveIndex], level) {
					moveIndex++
				}
			}
			m.Path = m.Path[:0]
		}
	}
}

func (m *Monster) Move(to Pos, level *Level) bool {
	moved := false
	p := level.Player
	tile := *level.TileAtPos(to)
	if m.Hitpoints > 0 && m.AP >= float64(tile.Cost) {
		_, exists := level.Monsters[to]
		if !exists && to != p.Pos {
			delete(level.Monsters, m.Pos)
			level.Monsters[to] = m
			m.Pos = to
			fmt.Println("moved!")
			m.AP -= float64(tile.Cost)
			moved = true
		} else if m.dist(p.Pos) == 1 && m.AP >= 1 {
			fmt.Println(p.posToString(), m.posToString())
			events := Attack(&m.Character, &level.Player.Character)
			level.AddEvents(events...)
		}
	}

	return moved
}

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
					if p == player {
						fmt.Println("target found")
						return true
					}
					level.Debug[p] = true
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
