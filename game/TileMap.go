package game

func (l *Level) LoadTileMap() {
	l.TileMap = make(map[rune]Tile)
	tiles := []Tile{
		{
			Type:     "Wall",
			Rune:     StoneWall,
			Passable: false,
			HasFloor: false,
			Occupied: false,
		},
		{
			Name:     "Dirt Floor",
			Type:     "Floor",
			Rune:     DirtFloor,
			Passable: true,
			HasFloor: true,
			Occupied: false,
		},
		{
			Name:     "Closed Door",
			Type:     "Door",
			Rune:     ClosedDoor,
			Passable: false,
			HasFloor: true,
			Occupied: false,
		},
		{
			Name:     "Open Door",
			Type:     "Door",
			Rune:     OpenDoor,
			Passable: true,
			HasFloor: true,
			Occupied: false,
		},
		{ // I don't like this
			Type:     "Empty",
			Rune:     Empty,
			Passable: false,
			HasFloor: false,
			Occupied: false,
		},
	}

	for _, t := range tiles {
		l.TileMap[t.Rune] = t
	}
}

// setTileVariance is pretty much a debug function to add information to tiles in the level
func (l *Level) SetTileVariance(pos Pos) {

}

func (l *Level) TileAtPos(pos Pos) *Tile {
	return &l.World[pos.Y][pos.X]
}
