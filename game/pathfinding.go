package game

import (
	"math"
)

// Might be useful for flood effects and stuff
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

func (level *Level) bresenham(start Pos, end Pos) []Pos {
	var line []Pos
	steep := math.Abs(float64(end.Y-start.Y)) > math.Abs(float64(end.X-start.X))
	if steep {
		start.X, start.Y = start.Y, start.X
		end.X, end.Y = end.Y, end.X
	}

	deltaY := int(math.Abs(float64(end.Y - start.Y)))
	err := 0
	y := start.Y
	ystep := 1
	if start.Y >= end.Y {
		ystep = -1
	}

	if start.X > end.X {
		deltaX := start.X - end.X
		for x := start.X; x > end.X; x-- {
			var pos Pos
			if steep {
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			line = append(line, pos)
			if !canSeeThrough(level, pos) {
				break
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	} else {
		deltaX := end.X - start.X
		for x := start.X; x < end.X; x++ {
			var pos Pos
			if steep {
				pos = Pos{y, x}
			} else {
				pos = Pos{x, y}
			}
			line = append(line, pos)
			if !canSeeThrough(level, pos) {
				break
			}
			err += deltaY
			if 2*err >= deltaX {
				y += ystep
				err -= deltaX
			}
		}
	}
	return line
}

func (level *Level) astar(from, to Pos) (path []Pos, dist int, found bool) {
	// fmt.Printf("start: {%d, %d}\ngoal: {%d, %d}\n", from.X, from.Y, to.X, to.Y)
	edge := make(pqueue, 0, 8)
	edge = edge.push(from, 1)
	prevPos := make(map[Pos]Pos)
	prevPos[from] = from
	currentCost := make(map[Pos]int)
	currentCost[from] = 0

	var current Pos
	for {
		if len(edge) == 0 {
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
			// for _, pos := range path {
			// 	level.Debug[pos] = true
			// }

			return path, len(currentCost), true
		}

		for _, next := range getNeighbours(level, current) {
			t := level.TileAtPos(next)
			newCost := currentCost[current] + t.Cost
			_, exists := currentCost[next]
			if !exists || newCost < currentCost[next] {
				currentCost[next] = newCost
				xDist := int(math.Abs(float64(to.X - next.X)))
				yDist := int(math.Abs(float64(to.Y - next.Y)))
				priority := newCost + xDist + yDist
				edge = edge.push(next, priority)
				// level.Debug[next] = true
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

	if canWalk(level, u) {
		neighbours = append(neighbours, u)
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
