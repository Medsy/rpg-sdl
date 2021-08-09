package game

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'R', "Rat"}, 5, 5, 1.5, 0.0}}
}

func NewSpider(p Pos) *Monster {
	return &Monster{Character{Entity{p, 'S', "Spider"}, 7, 7, 1, 0.0}}
}

func (m *Monster) Update(level *Level) {
	playerPos := level.Player.Pos
	path, _, found := level.astar(m.Pos, playerPos)
	moveIndex := 1

	if found {
		m.AP += m.Speed
		apInt := int(m.AP)
		for i := 0; i < apInt; i++ {
			if moveIndex < len(path) {
				m.Move(path[moveIndex], level)
				moveIndex++
			}
		}
	}
}

func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to]
	if !exists && to != level.Player.Pos {
		level.TileAtPos(m.Pos).Passable = true
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
		m.AP--
	} else if to == level.Player.Pos {
		Attack(&m.Character, &level.Player.Character)
		if m.Hitpoints <= 0 { //TODO: review death
			delete(level.Monsters, m.Pos)
		}
	}
	level.TileAtPos(m.Pos).Passable = false
}
