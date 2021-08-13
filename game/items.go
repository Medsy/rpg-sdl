package game

type Item struct {
	Entity
}

func NewSword(p Pos) *Item {
	return &Item{Entity{p, 's', "Sword"}}
}

func NewHelmet(p Pos) *Item {
	return &Item{Entity{p, 'h', "Helmet"}}
}
