package game

func (l *Level) LoadTileMap() {
	l.TileMap = make(map[rune]Tile)
	tiles := []Tile{
		{
			Type: "Wall",
			Rune: StoneWall,
			Passable: false,
		},
		{
			Type: "Floor",
			Rune: DirtFloor,
			Passable: true,
		},
		{
			Type: "Door",
			Rune: ClosedDoor,
			Passable: true,
		},
		{
			Type: "Door",
			Rune: OpenDoor,
			Passable: true,
		},
		{
			Type: "Player",
			Rune: PlayerTile,
		},
		{
			Type: "Monster",
			Rune: Rat,
			Passable: true,
		},
		{
			Type: "Monster",
			Rune: Spider,
			Passable: true,
		},
		{ // I don't like this
			Type: "Empty",
			Rune: Empty,
			Passable: false,
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
	return &l.Zone[pos.Y][pos.X]
}