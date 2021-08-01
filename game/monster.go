package game

type Monster struct {
	rune      rune
	name      string
	hitpoints int
	strength  int
	speed     float64
}

func NewRat() *Monster {
	return &Monster{'R', "Rat", 5,5, 2.0}
}

func NewSpider() *Monster {
	return &Monster{'S', "Spider", 8, 7, 1.0}
}