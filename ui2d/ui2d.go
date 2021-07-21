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

const winWidth, winHeight = 1280, 720

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture
var textureIndex map[game.Tile][]sdl.Rect

var centerX, centerY int32 = -1, -1


type UI2d struct {
	MouseInput game.Pos
}

func loadTextureIndex() {
	textureIndex = make(map[game.Tile][]sdl.Rect)
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

		textureIndex[tileRune] = rects
	}
}

func init() {

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}

	window, err := sdl.CreateWindow("rpg-sdl", 200, 200,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	textureAtlas, err = img.LoadTexture(renderer, "ui2d/assets/tiles.png")
	if err != nil {
		panic(err)
	}
	loadTextureIndex()
}

func (ui *UI2d) Draw(level *game.Level) {
	rand.Seed(1992) // needs to be called everytime before rendering with random tiles

	if centerX == -1 && centerY == -1 {
		centerX = level.Player.X
		centerY = level.Player.Y
	}

	threshold := int32(5)
	if level.Player.X > centerX+threshold {
		centerX++
	}
	if level.Player.X < centerX-threshold {
		centerX--
	}
	if level.Player.Y > centerY+threshold {
		centerY++
	}
	if level.Player.Y < centerY-threshold {
		centerY--
	}

	game.OffsetX = (winWidth / 2) - (centerX * 32)
	game.OffsetY = (winHeight / 2) - (centerY * 32)
	renderer.Clear()
	drawFloor(level, game.OffsetX, game.OffsetY)
	drawLevel(level, game.OffsetX, game.OffsetY)

	// Player tile 13, 51
	renderer.Copy(textureAtlas, &sdl.Rect{X: 13 * 32, Y: 59 * 32, W: 32, H: 32}, &sdl.Rect{X: level.Player.X*32 + game.OffsetX, Y: level.Player.Y*32 + game.OffsetY, W: 32, H: 32})
	renderer.Present()

	sdl.Delay(10)
}

// drawLevel receives a level from the game and then renders the tiles row by row
// if floor only is true, all tiles that are not Empty are drawn as dirt floor
func drawLevel(level *game.Level, offsetX, offsetY int32) {
	for y, row := range level.Zone {
		for x, tile := range row {
			if tile != game.Empty {
				if tile == game.DirtFloor {continue}
				srcs := textureIndex[tile]
				src := srcs[rand.Intn(len(srcs))]
				dst := sdl.Rect{X: int32(x*32) + offsetX, Y: int32(y*32) + offsetY, W: 32, H: 32} // TODO: maybe add a util to build rects with a configurable spritesheet defaults eg x,y,w,h

				pos := game.Pos{int32(x),int32(y)}
				if level.Debug[pos] {
					textureAtlas.SetColorMod(128, 0, 0)
				} else {
					textureAtlas.SetColorMod(255,255, 255)
				}

				renderer.Copy(textureAtlas, &src, &dst)
			}
		}
	}
}

func drawFloor(level *game.Level, offsetX, offsetY int32) {
	for y, row := range level.Zone {
		for x, tile := range row {
			if tile == game.DirtFloor || tile == game.OpenDoor || tile == game.ClosedDoor {
				srcs := textureIndex[game.DirtFloor]
				src := srcs[rand.Intn(len(srcs))]
				dst := sdl.Rect{X: int32(x*32) + offsetX, Y: int32(y*32) + offsetY, W: 32, H: 32} // TODO: maybe add a util to build rects with a configurable spritesheet defaults eg x,y,w,h

				pos := game.Pos{int32(x),int32(y)}
				if level.Debug[pos] {
					textureAtlas.SetColorMod(128, 0, 0)
				} else {
					textureAtlas.SetColorMod(255,255, 255)
				}

				renderer.Copy(textureAtlas, &src, &dst)
			}
		}
	}
}

func (ui *UI2d) GetInput() *game.Input {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			return &game.Input{Typ: game.Quit}
		case *sdl.MouseButtonEvent:
			if e.State == sdl.PRESSED {
				mousePos := game.Pos{X: e.X, Y: e.Y}
				return &game.Input{
					Typ: game.Search,
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
				return &game.Input{Typ: game.Quit}
			case sdl.K_UP, sdl.K_w:
				return &game.Input{Typ: game.Up}
			case sdl.K_DOWN, sdl.K_s:
				return &game.Input{Typ: game.Down}
			case sdl.K_LEFT, sdl.K_a:
				return &game.Input{Typ: game.Left}
			case sdl.K_RIGHT, sdl.K_d:
				return &game.Input{Typ: game.Right}
			case sdl.K_SPACE:
				return &game.Input{Typ: game.Search}
			}
		}
	}
	return &game.Input{Typ: game.None}
}
