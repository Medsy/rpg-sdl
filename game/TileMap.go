package game

type Tile struct {
	Name         string
	Type         string
	Rune         rune
	Visible      bool
	Seen         bool
	BloodStained bool
	HasFloor     bool
	Cost         int
}

const (
	PlayerTile   rune = '@'
	StoneWall    rune = '#'
	DirtFloor    rune = '.'
	ClosedDoor   rune = '|'
	OpenDoor     rune = '/'
	GlassBlock	 rune = 'G'
	Rat          rune = 'R'
	Spider       rune = 'S'
	Water        rune = '~'
	BloodStained rune = 'b'
	UpStairs     rune = 'u'
	DownStairs   rune = 'd'
	Empty        rune = 0
)

func (l *Level) LoadTileMap() {
	l.TileMap = make(map[rune]Tile)
	tiles := []Tile{
		{
			Type:     "Wall",
			Rune:     StoneWall,
			HasFloor: false,
			Cost:     0,
		},
		{
			Name:     "Dirt Floor",
			Type:     "Floor",
			Rune:     DirtFloor,
			HasFloor: true,
			Cost:     1,
		},
		{
			Name:     "Closed Door",
			Type:     "ClosedDoor", // seems pretty dumb...
			Rune:     ClosedDoor,
			HasFloor: true,
			Cost:     2,
		},
		{
			Name:     "Open Door",
			Type:     "Door",
			Rune:     OpenDoor,
			HasFloor: true,
			Cost:     1,
		},
		{
			Name:     "Glass Block",
			Type:     "Glass",
			Rune:     GlassBlock,
			HasFloor: true,
			Cost:     1,
		},
		{
			Name:     "Upstairs",
			Type:     "Upstairs",
			Rune:     UpStairs,
			HasFloor: true,
			Cost:     1,
		},
		{
			Name:     "Downstairs",
			Type:     "Downstairs",
			Rune:     DownStairs,
			HasFloor: true,
			Cost:     1,
		},
		{
			Name:     "Water",
			Type:     "Water",
			Rune:     Water,
			HasFloor: true,
			Cost:     2,
		},
		{ // I don't like this
			Type:     "Empty",
			Rune:     Empty,
			HasFloor: false,
			Cost:     0,
		},
	}

	for _, t := range tiles {
		l.TileMap[t.Rune] = t
	}
}

func (l *Level) TileAtPos(pos Pos) *Tile {
	// TODO: hold x y max on level
	if (pos.Y <= len(l.Level) || pos.X <= len(l.Level)) && (pos.X >= 0 || pos.Y >= 0) {
		return &l.Level[pos.Y][pos.X]
	}
	return &Tile{Type: "Empty"}
}
