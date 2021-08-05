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
	textureIndex map[rune][]sdl.Rect
	centerX      int32
	centerY      int32
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

func (ui *ui) QuitSDL() {
	sdl.Quit()
	ui.window.Destroy()
	ui.renderer.Destroy()
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[rune][]sdl.Rect)
	file, err := os.Open("ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0])

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

	game.OffsetX = int32(ui.winWidth/2) - (ui.centerX * 32)
	game.OffsetY = int32(ui.winHeight/2) - (ui.centerY * 32)
	ui.renderer.Clear()
	ui.drawFloor(level, game.OffsetX, game.OffsetY)
	ui.drawLevel(level, game.OffsetX, game.OffsetY)

	// Player tile 13, 51
	ui.renderer.Copy(ui.textureAtlas, &sdl.Rect{X: 13 * 32, Y: 59 * 32, W: 32, H: 32}, &sdl.Rect{X: level.Player.X*32 + game.OffsetX, Y: level.Player.Y*32 + game.OffsetY, W: 32, H: 32})
	ui.renderer.Present()
}

// drawLevel receives a level from the game and then renders the tiles row by row
// if floor only is true, all tiles that are not Empty are drawn as dirt floor
func (ui *ui) drawLevel(level *game.Level, offsetX, offsetY int32) {
	var currentPos game.Pos
	for y, row := range level.Zone {
		for x, tile := range row { // TODO: replace tile
			currentPos = game.Pos{X: int32(x), Y: int32(y)}
			if tile.Rune != game.Empty {
				if tile.Rune == game.DirtFloor {
					continue
				}
				srcs := ui.textureIndex[tile.Rune]
				v := rand.Intn(len(srcs))
				src := srcs[v]
				level.TileAtPos(currentPos).Variance = v // this is super unnecessary, either remove the variance or find a better way of setting it.
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
	var currentPos game.Pos
	for y, row := range level.Zone {
		for x, tile := range row {
			if tile.Type == "Floor" || tile.Type == "Door" {
				currentPos = game.Pos{X: int32(x), Y: int32(y)}
				srcs := ui.textureIndex[game.DirtFloor]
				v := rand.Intn(len(srcs))
				src := srcs[v]
				level.TileAtPos(currentPos).Variance = v // this is super unnecessary, either remove the variance or find a better way of setting it.
				dst := sdl.Rect{X: int32(x*32) + offsetX, Y: int32(y*32) + offsetY, W: 32, H: 32}
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
			var input game.Input
			switch e := event.(type) {
			case *sdl.QuitEvent:
				input.Type = game.CloseWindow
				input.LevelChan = ui.levelChan
				ui.QuitSDL()
				return
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					input.Type = game.CloseWindow
					input.LevelChan = ui.levelChan
					ui.QuitSDL()
					return
				}
			case *sdl.MouseButtonEvent:
				if e.State == sdl.PRESSED {
					if e.Button == sdl.BUTTON_LEFT {
						mousePos := game.Pos{X: e.X, Y: e.Y}
						ui.inputChan <- &game.Input{
							Type:     game.Search,
							MousePos: mousePos,
						}
					}
					if e.Button == sdl.BUTTON_RIGHT {
						mousePos := game.Pos{X: e.X, Y: e.Y}
						ui.inputChan <- &game.Input{
							Type:     game.Inspect,
							MousePos: mousePos,
						}
					}
				}
			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					break
				}
				var key sdl.Keycode
				switch key = e.Keysym.Sym; key {
				case sdl.K_ESCAPE:
					input.Type = game.CloseWindow
					input.LevelChan = ui.levelChan
					ui.QuitSDL()
					fmt.Println("quit")
					return
				case sdl.K_UP, sdl.K_w:
					input.Type = game.Up
					fmt.Println("up")
				case sdl.K_DOWN, sdl.K_s:
					input.Type = game.Down
					fmt.Println("down")
				case sdl.K_LEFT, sdl.K_a:
					input.Type = game.Left
					fmt.Println("left")
				case sdl.K_RIGHT, sdl.K_d:
					input.Type = game.Right
					fmt.Println("right")
				case sdl.K_SPACE:
					input.Type = game.Search
					fmt.Println("space")
				}
			}
			if input.Type != game.None {
				ui.inputChan <- &input
			}
		}
		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}
		sdl.Delay(16)
	}
}
