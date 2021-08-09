package game

import (
	"fmt"
	"math"
)

type Tile struct {
	Name     string
	Type     string
	Rune     rune
	Passable bool
	Occupied bool
	HasFloor bool
	Variance int
}

const (
	PlayerTile rune = '@'
	StoneWall  rune = '#'
	DirtFloor  rune = '.'
	ClosedDoor rune = '|'
	OpenDoor   rune = '/'
	Rat        rune = 'R'
	Spider     rune = 'S'
	Empty      rune = 0
)

func (level *Level) bfsearch(start Pos) {
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
			}
		}
	}
}

func (level *Level) astar(from, to Pos) (path []Pos, dist int, found bool) {
	// fmt.Printf("start: {%d, %d}\ngoal: {%d, %d}\n", from.X, from.Y, to.X, to.Y)
	edge := make(pqueue, 0, 8)
	edge = edge.push(from, 1)
	prevPos := make(map[Pos]Pos)
	prevPos[from] = from
	currentCost := make(map[Pos]int)
	currentCost[from] = 0

	level.Debug = make(map[Pos]bool)

	var current Pos
	for {
		if len(edge) == 0 {
			fmt.Println("no path")
			return
		}

		edge, current = edge.pop()
		if current == to {
			path := make([]Pos, 0)
			p := current
			for p != from {
				path = append(path, p)
				p = prevPos[p]
			}
			path = append(path, p)
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			level.Debug = make(map[Pos]bool)
			for _, pos := range path {
				level.Debug[pos] = true
			}

			return path, len(currentCost), true
		}

		for _, next := range getNeighbours(level, current) {
			var cost = 1
			// t := level.TileAtPos(next)
			// if t.Rune == ClosedDoor {
			// 	cost = 4
			// }
			newCost := currentCost[current] + cost
			_, exists := currentCost[next]
			if !exists || newCost < currentCost[next] {
				currentCost[next] = newCost
				xDist := int(math.Abs(float64(to.X - next.X)))
				yDist := int(math.Abs(float64(to.Y - next.Y)))
				priority := newCost + xDist + yDist
				edge = edge.push(next, priority)
				level.Debug[next] = true
				prevPos[next] = current
				//fmt.Printf("{%d, %d} to {%d, %d} cost: %d\n",current.X, current.Y, next.X, next.Y, newCost)
			}
		}
	}
}

func getNeighbours(level *Level, pos Pos) []Pos {
	neighbours := make([]Pos, 0, 4)
	u := Pos{pos.X, pos.Y - 1}
	d := Pos{pos.X, pos.Y + 1}
	l := Pos{pos.X - 1, pos.Y}
	r := Pos{pos.X + 1, pos.Y}

	if level.TileAtPos(u).Passable {
		neighbours = append(neighbours, u)
	}

	if level.TileAtPos(d).Passable {
		neighbours = append(neighbours, d)
	}

	if level.TileAtPos(l).Passable {
		neighbours = append(neighbours, l)
	}

	if level.TileAtPos(r).Passable {
		neighbours = append(neighbours, r)
	}

	return neighbours
}
