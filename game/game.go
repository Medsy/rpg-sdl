package game

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type Game struct {
	LevelChans []chan *Level
	InputChan  chan *Input
	Level      *Level
}

type InputType int

const (
	None InputType = iota
	Up
	Down
	Left
	Right
	Camera_Up
	Camera_Down
	Camera_Left
	Camera_Right
	Search
	Inspect
	QuitGame
	CloseWindow

	BloodVariantCount int = 12 // maybe I can get this from the atlas?
)

type Input struct {
	Type      InputType
	MousePos  Pos
	LevelChan chan *Level
}

var OffsetX, OffsetY int

type Pos struct {
	X, Y int
}

type Entity struct {
	Pos
	Rune rune
	Name string
}

type Character struct {
	Entity
	Type       string
	Hitpoints  int
	Strength   int
	Speed      float64
	SightRange int
	AP         float64
	Alive      bool
}

type Player struct {
	Character
}

type Level struct {
	Level    [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster
	Events   []string
	EventPos int
	TileMap  map[rune]Tile
	Debug    map[Pos]bool
	R        *rand.Rand
}

type priorityPos struct {
	Pos
	priority int
}

func NewGame(numWindows int, path string) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)
	}
	inputChan := make(chan *Input)

	return &Game{LevelChans: levelChans, InputChan: inputChan, Level: loadLevelFromFile(path)}
}

func loadLevelFromFile(filename string) *Level {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	level := &Level{}
	level.Debug = make(map[Pos]bool)
	level.Events = make([]string, 10)
	level.R = rand.New(rand.NewSource(1))
	scanner := bufio.NewScanner(file)
	zoneRows := make([]string, 0)
	longestRow := 0
	index := 0

	for scanner.Scan() {
		zoneRows = append(zoneRows, scanner.Text())
		if len(zoneRows[index]) > longestRow {
			longestRow = len(zoneRows[index])
		}
		index++
	}
	level.Level = make([][]Tile, len(zoneRows))

	level.Player = &Player{
		Character: Character{
			Entity: Entity{
				Pos:  Pos{},
				Rune: PlayerTile,
				Name: "meds",
			},
			Type:       "Player",
			Hitpoints:  20,
			Strength:   3,
			Speed:      1.0,
			SightRange: 7,
			Alive:      true,
			AP:         0,
		},
	}

	level.Monsters = make(map[Pos]*Monster)
	level.LoadTileMap()

	for i := range level.Level {
		level.Level[i] = make([]Tile, longestRow)
	}
	for y := 0; y < len(level.Level); y++ {
		line := zoneRows[y]
		for x, r := range line {
			var t Tile
			switch r {
			case ' ', '\n', '\t', '\r':
				t = level.TileMap[r]

			case '#':
				t = level.TileMap[r]
			case '.':
				t = level.TileMap[r]
			case '|':
				t = level.TileMap[r]
			case '/':
				t = level.TileMap[r]
			case '@':
				level.Player.X = x
				level.Player.Y = y
				t = level.TileMap[DirtFloor]
			case 'R':
				p := Pos{X: x, Y: y}
				level.Monsters[p] = NewRat(p)
				t = level.TileMap[DirtFloor]
			case 'S':
				p := Pos{X: x, Y: y}
				level.Monsters[p] = NewSpider(p)
				t = level.TileMap[DirtFloor]
			case '~':
				t = level.TileMap[r]
			default:
				panic(fmt.Sprintf("Invalid rune '%s' in map at position [%d,%d]", string(r), y+1, x+1))
			}
			level.Level[y][x] = t
		}
	}
	level.lineOfSight()
	return level
}

func canWalk(level *Level, pos Pos) bool {
	t := level.Level[pos.Y][pos.X]
	switch t.Type {
	case "Wall", "ClosedDoor", "Empty":
		return false
	}

	_, exists := level.Monsters[pos]
	return !exists
}

func canSeeThrough(level *Level, pos Pos) bool {
	if inRange(level, pos) {
		t := level.Level[pos.Y][pos.X]
		switch t.Type {
		case "Wall", "ClosedDoor", "Empty":
			return false
		default:
			return true
		}
	}
	return false
}

func (level *Level) lineOfSight() {
	pos := level.Player.Pos
	dist := level.Player.SightRange

	for y := pos.Y - dist; y <= pos.Y+dist; y++ {
		for x := pos.X - dist; x <= pos.X+dist; x++ {
			xDelta := pos.X - x
			yDelta := pos.Y - y
			d := math.Sqrt(float64(xDelta*xDelta + yDelta*yDelta))
			if d <= float64(dist) {
				level.bresenham(pos, Pos{x, y})
			}
		}
	}
}

func openDoor(level *Level, pos Pos) {
	t := level.Level[pos.Y][pos.X]
	if t.Rune == ClosedDoor {
		level.Level[pos.Y][pos.X] = level.TileMap[OpenDoor]
	}
	level.lineOfSight()
}

func (p *Player) Move(level *Level, to Pos) {
	_, exists := level.Monsters[to]
	if !exists {
		p.Pos = to
		for x, row := range level.Level {
			for y, _ := range row {
				level.Level[x][y].Visible = false
			}
		}
		level.lineOfSight()
	}
}

func (p *Player) Action(level *Level, target Pos) {

	t := level.TileAtPos(target)
	if t.Name == "Closed Door" {
		openDoor(level, target)
	}

	m, exists := level.Monsters[target]
	if exists {
		events := Attack(&p.Character, &m.Character)
		level.AddEvents(events...)

		if p.Hitpoints <= 0 {
			level.AddEvents("you died")
			p.Alive = false
		}
	}
}

func inRange(level *Level, pos Pos) bool {
	return int(pos.X) < len(level.Level[0]) && int(pos.Y) < len(level.Level) && pos.X >= 0 && pos.Y >= 0
}

func (g *Game) handleInput(input *Input) {
	level := g.Level
	p := level.Player
	switch input.Type {
	case Up:
		pos := Pos{p.X, p.Y - 1}
		if canWalk(level, pos) {
			level.Player.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Down:
		pos := Pos{p.X, p.Y + 1}
		if canWalk(level, pos) {
			level.Player.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Left:
		pos := Pos{p.X - 1, p.Y}
		if canWalk(level, pos) {
			level.Player.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Right:
		pos := Pos{p.X + 1, p.Y}
		if canWalk(level, pos) {
			level.Player.Move(level, pos)
		} else {
			p.Action(level, Pos{p.X + 1, p.Y})
		}
	case Camera_Up:

	case Camera_Down:

	case Camera_Left:

	case Camera_Right:

	case Search:
		target := screenToWorldPos(input.MousePos)
		t := level.TileAtPos(target)
		if t.Rune == DirtFloor {
		} else if t.Rune == ClosedDoor {
			openDoor(level, target)
		}
	case Inspect:
		tartget := screenToWorldPos(input.MousePos)
		t := level.TileAtPos(tartget)
		if t.HasFloor {
			spew.Dump(t)
		}
		m, exists := level.Monsters[tartget]
		if exists {
			spew.Dump(m)
		}
	case CloseWindow:
		close(input.LevelChan)
		chanIndex := 0
		for i, c := range g.LevelChans {
			if c == input.LevelChan {
				chanIndex = i
				break
			}
		}
		g.LevelChans = append(g.LevelChans[:chanIndex], g.LevelChans[chanIndex+1:]...)
		g.LevelChans = append(g.LevelChans[:chanIndex], g.LevelChans[chanIndex+1:]...)
	case None:
		break
	}
}

func (level *Level) AddEvents(events ...string) {
	for _, event := range events {
		level.Events[level.EventPos] = event

		level.EventPos++
		if level.EventPos == len(level.Events) {
			level.EventPos = 0
		}

	}
}

func screenToWorldPos(pos Pos) Pos {
	return Pos{(pos.X - OffsetX) / 32, (pos.Y - OffsetY) / 32}
}

var eventCount int

func (g *Game) Run() {
	fmt.Println("Starting...")

	for _, lchan := range g.LevelChans {
		lchan <- g.Level
	}
	count := 1
	for _, m := range g.Level.Monsters {
		m.Name = m.Name + " " + fmt.Sprint(count)
		count++
	}

	for input := range g.InputChan {
		if input.Type == QuitGame {
			return
		}
		g.Level.Debug = map[Pos]bool{}

		if g.Level.Player.Alive {
			for _, monster := range g.Level.Monsters {
				monster.Update(g.Level)
			}

			g.handleInput(input)
		}

		if len(g.LevelChans) == 0 {
			return
		}
		for _, lchan := range g.LevelChans {
			lchan <- g.Level
		}
	}
}
