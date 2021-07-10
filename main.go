package main

import (
	"rpg-sdl/game"
	"rpg-sdl/ui2d"
)

func main() {
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
