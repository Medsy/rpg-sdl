package game

import (
	"bufio"
	"fmt"
	"os"
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
)

type Input struct {
	Type      InputType
	MousePos  Pos
	LevelChan chan *Level
}

var OffsetX int32
var OffsetY int32

type Pos struct {
	X, Y int32
}

type Entity struct {
	Pos
	Rune rune
	Name string
}

type Character struct {
	Entity
	Hitpoints int
	Strength  int
	Speed     float64
	AP        float64
}

type Player struct {
	Character
}

type Level struct {
	World    [][]Tile
	Player   *Player
	Monsters map[Pos]*Monster
	TileMap  map[rune]Tile
	Debug    map[Pos]bool
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
	level.World = make([][]Tile, len(zoneRows))

	level.Player = &Player{
		Character: Character{
			Entity: Entity{
				Pos:  Pos{},
				Rune: '@',
				Name: "meds",
			},
			Hitpoints: 20,
			Strength:  20,
			Speed:     1.0,
			AP:        0,
		},
	}

	level.Monsters = make(map[Pos]*Monster)
	level.LoadTileMap()

	for i := range level.World {
		level.World[i] = make([]Tile, longestRow)
	}
	for y := 0; y < len(level.World); y++ {
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
				level.Player.X = int32(x)
				level.Player.Y = int32(y)
				t = level.TileMap[DirtFloor]
			case 'R':
				p := Pos{X: int32(x), Y: int32(y)}
				level.Monsters[p] = NewRat(p)
				t = level.TileMap[DirtFloor]
			case 'S':
				p := Pos{X: int32(x), Y: int32(y)}
				level.Monsters[p] = NewSpider(p)
				t = level.TileMap[DirtFloor]
			default:
				panic(fmt.Sprintf("Invalid rune '%s' in map at position [%d,%d]", string(r), y+1, x+1))
			}
			level.World[y][x] = t
		}
	}

	return level
}

func canWalk(level *Level, pos Pos) bool {
	t := level.World[pos.Y][pos.X]
	return t.Passable
}

func openDoor(level *Level, pos Pos) {
	t := level.World[pos.Y][pos.X]
	if t.Rune == ClosedDoor {
		level.World[pos.Y][pos.X] = level.TileMap[OpenDoor]
	}
}

func (p *Player) Move(level *Level, to Pos) {
	_, exists := level.Monsters[to]
	if !exists {
		p.Pos = to
	}
}

func (p *Player) Action(level *Level, target Pos) {
	m, exists := level.Monsters[target]
	if exists {
		Attack(&p.Character, &m.Character)
		if m.Hitpoints <= 0 {
			level.TileAtPos(m.Pos).Passable = true
			delete(level.Monsters, m.Pos)
		}
		if p.Hitpoints <= 0 {
			panic("You Died")
		}
	}

	t := level.TileAtPos(target)
	if t.Name == "Closed Door" {
		openDoor(level, target)
	}
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
			level.astar(p.Pos, target)
		} else if t.Rune == ClosedDoor {
			openDoor(level, target)
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

func screenToWorldPos(pos Pos) Pos {
	return Pos{(pos.X - OffsetX) / 32, (pos.Y - OffsetY) / 32}
}

func (g *Game) Run() {
	fmt.Println("Starting...")

	for _, lchan := range g.LevelChans {
		lchan <- g.Level
	}

	for input := range g.InputChan {
		if input.Type == QuitGame {
			return
		}
		g.handleInput(input)

		for _, monster := range g.Level.Monsters {
			monster.Update(g.Level)
		}

		if len(g.LevelChans) == 0 {
			return
		}
		for _, lchan := range g.LevelChans {
			lchan <- g.Level
		}
	}
}
