package game

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
)

type GameUI interface {
	Draw(*Level)
	GetInput() *Input
}

type InputType int
const (
	None InputType = iota
	Up
	Down
	Left
	Right
	Search
	Quit
)

type Input struct {
	Typ InputType
}

type Tile rune

const (
	Character1	Tile = 'P'
	StoneWall   Tile = '#'
	DirtFloor   Tile = '.'
	ClosedDoor  Tile = '|'
	OpenDoor  	Tile = '/'
	Empty       Tile = 0
)

type Pos struct {
	X, Y int32
}

type Entity struct {
	Pos
}

type Player struct {
	Entity
}

type Level struct {
	Zone   [][]Tile
	Player Player
	Debug  map[Pos]bool
}

type priorityPos struct {
	Pos
	priority int
}

type priorityArray []priorityPos
func (p priorityArray) Len() int {return len(p)}
func (p priorityArray) Swap(i, j int) {p[i], p[j] = p[j], p[i]}
func (p priorityArray) Less(i, j int) bool { return p[i].priority < p[j].priority}


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
	level.Zone = make([][]Tile, len(zoneRows))

	for i := range level.Zone {
		level.Zone[i] = make([]Tile, longestRow)
	}
	for y := 0; y < len(level.Zone); y++ {
		line := zoneRows[y]
		for x, r := range line {
			var t Tile
			switch r {
			case ' ', '\n', '\t', '\r':
				t = Empty
			case '#':
				t = StoneWall
			case '.':
				t = DirtFloor
			case '|':
				t = ClosedDoor
			case '/':
				t = OpenDoor
			case 'P':
				level.Player.X = int32(x)
				level.Player.Y = int32(y)
				t = DirtFloor
			default:
				panic(fmt.Sprintf("Invalid rune '%s' in map at position [%d,%d]", string(r), y+1, x+1))
			}
			level.Zone[y][x] = t
		}
	}

	return level
}

func canWalk(level *Level, pos Pos) bool {
	t := level.Zone[pos.Y][pos.X]
	switch t {
	case StoneWall, ClosedDoor, Empty:
		return false
	default:
		return true
	}
}

func checkDoor(level *Level, pos Pos) {
	t := level.Zone[pos.Y][pos.X]
	if t == ClosedDoor {
		level.Zone[pos.Y][pos.X] = OpenDoor
	}
}

func handleInput(ui GameUI, level *Level, input *Input)  {
	p := level.Player
	switch input.Typ {
	case Up:
		if canWalk(level, Pos{p.X, p.Y-1}){
			level.Player.Y--
		} else {
			checkDoor(level, Pos{p.X, p.Y-1})
		}
	case Down:
		if canWalk(level, Pos{p.X, p.Y+1}){
			level.Player.Y++
		} else {
			checkDoor(level, Pos{p.X, p.Y+1})
		}
	case Left:
		if canWalk(level, Pos{p.X-1, p.Y}){
			level.Player.X--
		} else {
			checkDoor(level, Pos{p.X-1, p.Y})
		}
	case Right:
		if canWalk(level, Pos{p.X+1, p.Y}){
			level.Player.X++
		} else {
			checkDoor(level, Pos{p.X+1, p.Y})
		}
	case Search:
		astar(ui, level, p.Pos, Pos{16 ,15})
	case None:
		break
	}
}

func bfsearch(ui GameUI, level *Level, start Pos) {
	edge := make([]Pos, 0, 8)
	edge = append(edge, start)
	visited := make(map[Pos]bool)
	visited[start] = true
	level.Debug = visited

	for len(edge) > 0 {
		current := edge[0]
		edge = edge[1:]
		for _, next := range getNeighbours(level, current) {
			if !visited[next] {
				edge = append(edge, next)
				visited[next] = true
				ui.Draw(level)
			}
		}
	}
}

func getNeighbours(level *Level, pos Pos) []Pos {
		neighbours := make([]Pos,0,4)
		u := Pos{pos.X, pos.Y-1}
		d := Pos{pos.X, pos.Y+1}
		l := Pos{pos.X-1, pos.Y}
		r := Pos{pos.X+1, pos.Y}

		if canWalk(level, u) {
			neighbours = append(neighbours,u)
		}
		if canWalk(level, d) {
			neighbours = append(neighbours, d)
		}
		if canWalk(level, l) {
			neighbours = append(neighbours, l)
		}
		if canWalk(level, r) {
			neighbours = append(neighbours, r)
		}

		return neighbours
}

func astar(ui GameUI, level *Level, start Pos, goal Pos) []Pos {
	edge := make(priorityArray, 0, 8)
	edge = append(edge, priorityPos{start, 1})
	prevPos := make(map[Pos]Pos)
	prevPos[start] = start
	currentCost := make(map[Pos]int)
	currentCost[start] = 0

	level.Debug = make(map[Pos]bool)

	for len(edge) > 0 {
		sort.Stable(edge)
		current := edge[0]

		if current.Pos == goal {
			path := make([]Pos, 0)
			p := current.Pos
			for p != start {
				path = append(path, p)
				p = prevPos[p]
			}
			path = append(path, p)
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}

			for _, pos := range path {
				level.Debug[pos] = true
				ui.Draw(level)
				ui.Draw(level)
			}

			return path
			break
		}

		edge = edge[1:]
		for _, next := range getNeighbours(level, current.Pos) {
			newCost := currentCost[current.Pos] + 1
			_, exists := currentCost[next]
			if !exists || newCost < currentCost[next] {
				currentCost[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				edge = append(edge, priorityPos{next, priority})
				prevPos[next] = current.Pos

			}
		}
	}

	return nil
}

func Run(ui GameUI) {
	level := loadLevelFromFile("game/maps/level1.map")
	for {
		ui.Draw(level)
		input := ui.GetInput()
		if input != nil && input.Typ == Quit{
			return
		}

		handleInput(ui, level, input)
	}
}
