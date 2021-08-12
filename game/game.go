package game

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

type Game struct {
	LevelChans   []chan *Level
	InputChan    chan *Input
	Levels       map[string]*Level
	CurrentLevel *Level
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
	StairMap map[Pos]*LevelPos
	Events   []string
	EventPos int
	TileMap  map[rune]Tile
	Debug    map[Pos]bool
	R        *rand.Rand
}

type LevelPos struct {
	Level *Level
	Pos
}

type priorityPos struct {
	Pos
	priority int
}

func NewGame(numWindows int) *Game {
	levelChans := make([]chan *Level, numWindows)
	for i := range levelChans {
		levelChans[i] = make(chan *Level)

	}
	inputChan := make(chan *Input)
	//TODO: need to better select the first level

	game := &Game{LevelChans: levelChans, InputChan: inputChan, Levels: loadLevels()}
	game.loadWorld()

	return game
}

func (p *Pos) posToString() string {
	return fmt.Sprintf("{%d, %d}", p.X, p.Y)
}

func (game *Game) loadWorld() {
	file, err := os.Open("game/maps/world.txt")
	if err != nil {
		panic(err)
	}
	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	rows, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	for rowIndex, row := range rows {
		// set first level
		if rowIndex == 0 {
			game.CurrentLevel = game.Levels[row[0]]
			if game.CurrentLevel == nil {
				fmt.Println("couldn't find currentlevel name in world file:", row[0])
				panic(nil)
			}
			continue
		}
		levelWithStairs := game.Levels[row[0]]
		if levelWithStairs == nil {
			fmt.Println("couldn't find level name 1 in world file")
			panic(nil)
		}
		x, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			panic(err)
		}
		pos := Pos{int(x), int(y)}

		levelToTeleportTo := game.Levels[row[3]]
		if levelWithStairs == nil {
			fmt.Println("couldn't find level name 2 in world file")
			panic(nil)
		}
		x, err = strconv.ParseInt(row[4], 10, 64)
		if err != nil {
			panic(err)
		}
		y, err = strconv.ParseInt(row[5], 10, 64)
		if err != nil {
			panic(err)
		}
		posToClimbTo := Pos{int(x), int(y)}
		levelWithStairs.StairMap[pos] = &LevelPos{levelToTeleportTo, posToClimbTo}
	}
}

func loadLevels() map[string]*Level {
	newPlayer := &Player{
		Character: Character{
			Entity: Entity{
				Pos:  Pos{},
				Rune: PlayerTile,
				Name: "meds",
			},
			Type:       "Player",
			Hitpoints:  50,
			Strength:   3,
			Speed:      1.0,
			SightRange: 5,
			Alive:      true,
			AP:         0,
		},
	}

	levels := make(map[string]*Level)

	filenames, err := filepath.Glob("game/maps/*.map")
	if err != nil {
		panic(err)
	}
	for _, fileName := range filenames {

		levelName := fileName[len(filepath.Base(fileName)) : len(fileName)-len(filepath.Ext(fileName))] // Readable? Nah... It just extracts the filename from the path
		file, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		level := &Level{}
		level.Debug = make(map[Pos]bool)
		level.Events = make([]string, 10)
		level.R = rand.New(rand.NewSource(1))
		level.StairMap = make(map[Pos]*LevelPos)
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

		level.Player = newPlayer

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
					t = level.TileMap[Empty]
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
				case 'u':
					t = level.TileMap[r]
				case 'd':
					t = level.TileMap[r]
				default:
					panic(fmt.Sprintf("Invalid rune '%s' in map at position [%d,%d]", string(r), y+1, x+1))
				}
				level.Level[y][x] = t
			}
		}
		levels[levelName] = level
	}
	return levels
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
				zone := level.bresenham(pos, Pos{x, y})
				for _, pos := range zone {
					level.TileAtPos(pos).Visible = true
					level.TileAtPos(pos).Seen = true
				}
			}
		}
	}
}

func openDoor(level *Level, pos Pos) {
	t := level.Level[pos.Y][pos.X]
	if t.Rune == ClosedDoor {
		level.Level[pos.Y][pos.X] = level.TileMap[OpenDoor]
		level.lineOfSight()
		level.Player.AP--
	}

}

func (g *Game) Move(level *Level, to Pos) {
	if inRange(g.CurrentLevel, to) {
		p := level.Player
		tile := level.TileAtPos(to)
		if p.AP >= float64(tile.Cost) {
			stairs := level.StairMap[to]
			if stairs != nil {
				g.CurrentLevel = stairs.Level
				g.CurrentLevel.Player.Pos = stairs.Pos
				p.AP -= float64(tile.Cost)
				g.CurrentLevel.lineOfSight()
			} else {
				_, exists := level.Monsters[to]
				if !exists {
					p.Pos = to
					p.AP -= float64(tile.Cost)
					for x, row := range level.Level {
						for y := range row {
							level.Level[x][y].Visible = false
						}
					}
					level.lineOfSight()
				}
			}
		}
	}
}

// pass game to these fuctions and cut down params
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
	level := g.CurrentLevel
	p := level.Player
	switch input.Type {
	case Up:
		pos := Pos{p.X, p.Y - 1}
		if canWalk(level, pos) {
			g.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Down:
		pos := Pos{p.X, p.Y + 1}
		if canWalk(level, pos) {
			g.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Left:
		pos := Pos{p.X - 1, p.Y}
		if canWalk(level, pos) {
			g.Move(level, pos)
		} else {
			p.Action(level, pos)
		}
	case Right:
		pos := Pos{p.X + 1, p.Y}
		if canWalk(level, pos) {
			g.Move(level, pos)
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
		if t.Rune == ClosedDoor {
			openDoor(level, target)
		}
	case Inspect:
		target := screenToWorldPos(input.MousePos)
		spew.Dump(level.TileAtPos(target))

		m, exists := level.Monsters[target]
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

func (p *Player) GetStatStrings() []string {
	var stats []string

	stats = append(stats, "Name: "+p.Name)
	stats = append(stats, "HP: "+fmt.Sprint(p.Hitpoints))
	stats = append(stats, "Str: "+fmt.Sprint(p.Strength))
	stats = append(stats, "Spd: "+fmt.Sprint(int(p.Speed)))
	stats = append(stats, "AP: "+fmt.Sprint(int(p.AP)))
	stats = append(stats, "Pos: "+p.posToString())

	return stats
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

//  apparently ambit means range
func (e *Entity) InRange(ambit int, p Pos) bool {
	dist := int(math.Abs(float64(p.X-e.X) + float64(e.Y-p.Y)))
	fmt.Printf("e.Pos: %s p.Pos: %s\n", e.posToString(), p.posToString())
	fmt.Printf("formula: (%d - %d) + (%d - %d)\n", e.X, p.X, e.Y, p.Y)
	fmt.Printf("calc dist: %d\n", dist)
	fmt.Printf("in range: %v\n", dist <= ambit)
	return dist <= ambit
}

func (g *Game) Run() {
	fmt.Println("Starting...")

	for _, lchan := range g.LevelChans {
		lchan <- g.CurrentLevel
	}
	count := 1
	for _, m := range g.CurrentLevel.Monsters {
		m.Name = m.Name + " " + fmt.Sprint(count)
		count++
	}

	g.CurrentLevel.lineOfSight()

	for input := range g.InputChan {
		if input.Type == QuitGame {
			return
		}
		g.CurrentLevel.Debug = map[Pos]bool{}

		if g.CurrentLevel.Player.Alive {
			for _, monster := range g.CurrentLevel.Monsters {
				if monster.isPlayerInRange(g.CurrentLevel) {
					fmt.Println(monster.Name)
					fmt.Println("---------")
					monster.Update(g.CurrentLevel)
				}
			}

			g.handleInput(input)
		}

		if len(g.LevelChans) == 0 {
			return
		}
		for _, lchan := range g.LevelChans {
			lchan <- g.CurrentLevel
		}
		fmt.Println(g.CurrentLevel.Debug)
		g.CurrentLevel.Player.AP += g.CurrentLevel.Player.Speed
	}
}
