package ui2d

import (
	"rpg-sdl/game"

	"github.com/veandco/go-sdl2/sdl"
)

type mouseState struct {
	leftButton  bool
	rightButton bool
	pos         game.Pos
}

func getMouseState() *mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := leftButton & sdl.ButtonRMask()
	var result mouseState
	result.pos = game.Pos{int(mouseX), int(mouseY)}
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)

	return &result
}

func (ui *ui) keyDownOnce(key uint8) bool {
	return ui.keyboardState[key] == 1 && ui.prevKeyboardState[key] == 0
}

func (ui *ui) keyPressed(key uint8) bool {
	return ui.keyboardState[key] == 0 && ui.prevKeyboardState[key] == 1
}
