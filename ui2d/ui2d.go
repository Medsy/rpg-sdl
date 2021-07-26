package ui2d

import (
	"bufio"
	"fmt"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"math/rand"
	"os"
	"rpg-sdl/game"
	"strconv"
	"strings"
)

type ui struct {
	winWidth     int
	winHeight    int
	window       *sdl.Window
	renderer     *sdl.Renderer
	textureAtlas *sdl.Texture
	textureIndex map[game.Tile][]sdl.Rect
	centerX      int32
	centerY      int32
	offsetX      int32
	offsetY      int32
	levelChan    chan *game.Level
	inputChan    chan *game.Input
}

type UI2d struct {
	MouseInput game.Pos
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.winWidth = 1280
	ui.winHeight = 720

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}

	ui.window, err = sdl.CreateWindow("rpg-sdl", 200, 200,
		int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas, err = img.LoadTexture(ui.renderer, "ui2d/assets/tiles.png")
	if err != nil {
		panic(err)
	}
	ui.loadTextureIndex()

	ui.centerX = -1
	ui.centerY = -1

	return ui
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[game.Tile][]sdl.Rect)
	file, err := os.Open("ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := game.Tile(line[0])

		xy := line[1:]
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		if err != nil {
			panic(err)
		}

		var rects []sdl.Rect
		for i := 0; i < int(variationCount); i++ {
			rects = append(rects, sdl.Rect{X: int32(x * 32), Y: int32(y * 32), W: 32, H: 32})
			x++
			if x > 63 {
				x = 0
				y++
			}
		}

		ui.textureIndex[tileRune] = rects

	}
}

func (ui *ui) Draw(level *game.Level) {
	rand.Seed(1992) // needs to be called everytime before rendering with random tiles

	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}

	threshold := int32(5)
	if level.Player.X > ui.centerX+threshold {
		ui.centerX++
	}
	if level.Player.X < ui.centerX-threshold {
		ui.centerX--
	}
	if level.Player.Y > ui.centerY+threshold {
		ui.centerY++
	}
	if level.Player.Y < ui.centerY-threshold {
		ui.centerY--
	}

	ui.offsetX = int32(ui.winWidth/2) - (ui.centerX * 32)
	ui.offsetY = int32(ui.winHeight/2) - (ui.centerY * 32)
	ui.renderer.Clear()
	ui.drawFloor(level, ui.offsetX, ui.offsetY)
	ui.drawLevel(level, ui.offsetX, ui.offsetY)

	// Player tile 13, 51
	ui.renderer.Copy(ui.textureAtlas, &sdl.Rect{X: 13 * 32, Y: 59 * 32, W: 32, H: 32}, &sdl.Rect{X: level.Player.X*32 + ui.offsetX, Y: level.Player.Y*32 + ui.offsetY, W: 32, H: 32})
	ui.renderer.Present()

	sdl.Delay(10)
}

// drawLevel receives a level from the game and then renders the tiles row by row
// if floor only is true, all tiles that are not Empty are drawn as dirt floor
func (ui *ui) drawLevel(level *game.Level, offsetX, offsetY int32) {
	for y, row := range level.Zone {
		for x, tile := range row {
			if tile != game.Empty {
				if tile == game.DirtFloor {
					continue
				}
				srcs := ui.textureIndex[tile]
				src := srcs[rand.Intn(len(srcs))]
				dst := sdl.Rect{X: int32(x*32) + offsetX, Y: int32(y*32) + offsetY, W: 32, H: 32} // TODO: maybe add a util to build rects with a configurable spritesheet defaults eg x,y,w,h

				pos := game.Pos{int32(x), int32(y)}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0)
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}

				ui.renderer.Copy(ui.textureAtlas, &src, &dst)
			}
		}
	}
}

func (ui *ui) drawFloor(level *game.Level, offsetX, offsetY int32) {
	for y, row := range level.Zone {
		for x, tile := range row {
			if tile == game.DirtFloor || tile == game.OpenDoor || tile == game.ClosedDoor {
				srcs := ui.textureIndex[game.DirtFloor]
				src := srcs[rand.Intn(len(srcs))]
				dst := sdl.Rect{X: int32(x*32) + offsetX, Y: int32(y*32) + offsetY, W: 32, H: 32} // TODO: maybe add a util to build rects with a configurable spritesheet defaults eg x,y,w,h

				pos := game.Pos{int32(x), int32(y)}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0)
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}

				ui.renderer.Copy(ui.textureAtlas, &src, &dst)
			}
		}
	}
}

func (ui *ui) GetInput() {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChan: ui.levelChan}
				}
			case *sdl.MouseButtonEvent:
				if e.State == sdl.PRESSED {
					mousePos := game.Pos{X: e.X, Y: e.Y}
					ui.inputChan <- &game.Input{
						Typ:      game.Search,
						MousePos: mousePos,
					}
				}
			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					break
				}
				var key sdl.Keycode
				switch key = e.Keysym.Sym; key {
				case sdl.K_ESCAPE:
					ui.inputChan <- &game.Input{Typ: game.QuitGame}
					fmt.Println("quit")
				case sdl.K_UP, sdl.K_w:
					ui.inputChan <- &game.Input{Typ: game.Up}
					fmt.Println("up")
				case sdl.K_DOWN, sdl.K_s:
					ui.inputChan <- &game.Input{Typ: game.Down}
					fmt.Println("down")
				case sdl.K_LEFT, sdl.K_a:
					ui.inputChan <- &game.Input{Typ: game.Left}
					fmt.Println("left")
				case sdl.K_RIGHT, sdl.K_d:
					ui.inputChan <- &game.Input{Typ: game.Right}
					fmt.Println("right")
				case sdl.K_SPACE:
					ui.inputChan <- &game.Input{Typ: game.Search}
					fmt.Println("space")
				}
			}
		}
		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}
		ui.inputChan <- &game.Input{Typ: game.None}
	}
}
