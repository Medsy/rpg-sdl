package main

import (
	"rpg-sdl/game"
	"rpg-sdl/ui2d"
	"runtime"
)

func main() {
	runtime.LockOSThread()
	g := game.NewGame(1, "game/maps/level1.map")

	go func() {g.Run()}()
	ui := ui2d.NewUI(g.InputChan, g.LevelChans[0])
	ui.GetInput()
}
