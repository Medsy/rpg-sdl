package ui2d

import (
	"bufio"
	"math/rand"
	"os"
	"rpg-sdl/game"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ui struct {
	winWidth        int
	winHeight       int
	window          *sdl.Window
	renderer        *sdl.Renderer
	textureAtlas    *sdl.Texture
	fontSmall       *ttf.Font
	fontMedium      *ttf.Font
	fontLarge       *ttf.Font
	panelBackground *sdl.Texture
	textureIndex    map[rune][]sdl.Rect
	centerX         int
	centerY         int
	levelChan       chan *game.Level
	inputChan       chan *game.Input

	prevKeyboardState []uint8
	keyboardState     []uint8
	currentMouseState *mouseState
	prevMouseState    *mouseState

	r *rand.Rand

	strToTexSmall  map[string]*sdl.Texture
	strToTexMedium map[string]*sdl.Texture
	strToTexLarge  map[string]*sdl.Texture
	debug          bool
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func Init() {
	// sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}

	ui.debug = false

	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.strToTexSmall = make(map[string]*sdl.Texture)  // TODO: maybe prevent using 3 maps by combining the
	ui.strToTexMedium = make(map[string]*sdl.Texture) // string with the fontsize like
	ui.strToTexLarge = make(map[string]*sdl.Texture)  // "1:this is my string" with 1 meaning small
	ui.winWidth = 1280
	ui.winHeight = 720
	ui.r = rand.New(rand.NewSource(1))

	var err error
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

	if err := ttf.Init(); err != nil {
		panic(err)
	}

	ui.fontSmall, err = ttf.OpenFont("ui2d/assets/font.ttf", 16)
	if err != nil {
		panic(err)
	}

	ui.fontMedium, err = ttf.OpenFont("ui2d/assets/font.ttf", 24)
	if err != nil {
		panic(err)
	}

	ui.fontLarge, err = ttf.OpenFont("ui2d/assets/font.ttf", 32)
	if err != nil {
		panic(err)
	}

	ui.textureAtlas, err = img.LoadTexture(ui.renderer, "ui2d/assets/tiles.png")
	if err != nil {
		panic(err)
	}
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.prevKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyboardState[i] = v
	}

	ui.centerX = -1
	ui.centerY = -1

	ui.panelBackground = ui.GetSinglePixelTex(sdl.Color{0, 0, 0, 128})
	ui.panelBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	return ui
}

func (ui *ui) stringToTexture(string string, color sdl.Color, size FontSize) *sdl.Texture {
	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.strToTexSmall[string]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.strToTexMedium[string]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.strToTexLarge[string]
		if exists {
			return tex
		}
	}

	fontSurface, err := font.RenderUTF8BlendedWrapped(string, color, 512)
	if err != nil {
		panic(err)
	}

	tex, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}

	switch size {
	case FontSmall:
		ui.strToTexSmall[string] = tex
	case FontMedium:
		ui.strToTexMedium[string] = tex
	case FontLarge:
		ui.strToTexLarge[string] = tex
	}

	return tex
}

func (ui *ui) QuitSDL() {
	sdl.Quit()
	ui.window.Destroy()
	ui.renderer.Destroy()
	// ui.font.Close()
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
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 24)
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
	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}

	threshold := 5
	if level.Player.X > ui.centerX+threshold {
		diff := level.Player.X - (ui.centerX + threshold)
		ui.centerX += diff
	} else if level.Player.X < ui.centerX-threshold {
		diff := (ui.centerX - threshold) - level.Player.X
		ui.centerX -= diff
	} else if level.Player.Y > ui.centerY+threshold {
		diff := level.Player.Y - (ui.centerY + threshold)
		ui.centerY += diff
	} else if level.Player.Y < ui.centerY-threshold {
		diff := (ui.centerY - threshold) - level.Player.Y
		ui.centerY -= diff
	}

	game.OffsetX = (ui.winWidth / 2) - (ui.centerX * 32)
	game.OffsetY = (ui.winHeight / 2) - (ui.centerY * 32)
	ui.renderer.Clear()
	ui.r.Seed(63)

	ui.drawFloor(level, game.OffsetX, game.OffsetY)
	ui.drawLevel(level, game.OffsetX, game.OffsetY)
	ui.drawOnFloor(level, game.OffsetX, game.OffsetY)

	ui.textureAtlas.SetColorMod(255, 255, 255) // needed or sometimes entities stay modded

	for pos, monster := range level.Monsters {
		if level.Level[pos.Y][pos.X].Visible {
			monsterSrcRect := ui.textureIndex[monster.Rune][0]

			ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{X: int32(pos.X*32 + game.OffsetX), Y: int32(pos.Y*32 + game.OffsetY), W: 32, H: 32})
		}
	}

	// Player tile 13, 59
	playerSrcRect := ui.textureIndex[game.PlayerTile][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{X: int32(level.Player.X*32 + game.OffsetX), Y: int32(level.Player.Y*32 + game.OffsetY), W: 32, H: 32}) //TODO: custom rect builder

	ui.drawUI(level)

	ui.renderer.Present()
}

func (ui *ui) renderDebug(level *game.Level, pos game.Pos) {
	if ui.debug {
		if level.Debug[pos] {
			ui.textureAtlas.SetColorMod(128, 0, 0)
		} else {
			ui.textureAtlas.SetColorMod(255, 255, 255)
		}
	}
}

// drawLevel receives a level from the game and then renders the tiles row by row
// if floor only is true, all tiles that are not Empty are drawn as dirt floor
func (ui *ui) drawLevel(level *game.Level, offsetX, offsetY int) {

	for y, row := range level.Level {
		// loop over each tile per row
		for x, tile := range row {
			if tile.Rune != game.Empty {
				srcs := ui.textureIndex[tile.Rune]
				src := srcs[ui.r.Intn(len(srcs))]
				if tile.Visible || tile.Seen {
					if tile.Type == "Floor" || tile.Type == "Player" || tile.Type == "Monster" {
						continue
					}
					dst := sdl.Rect{X: int32(x*32 + offsetX), Y: int32(y*32 + offsetY), W: 32, H: 32} // TODO: maybe add a util to build rects with a configurable spritesheet defaults eg x,y,w,h
					pos := game.Pos{X: x, Y: y}
					if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else if tile.Visible {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}
					ui.renderDebug(level, pos)

					ui.renderer.Copy(ui.textureAtlas, &src, &dst)
				}
			}
		}
	}
}

func (ui *ui) drawFloor(level *game.Level, offsetX, offsetY int) {
	for y, row := range level.Level {
		for x, tile := range row {
			if tile.HasFloor {
				srcs := ui.textureIndex[game.DirtFloor]
				src := srcs[ui.r.Intn(len(srcs))]
				if tile.Visible || tile.Seen {
					dst := sdl.Rect{X: int32(x*32 + offsetX), Y: int32(y*32 + offsetY), W: 32, H: 32}
					pos := game.Pos{X: x, Y: y}
					if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else if tile.Visible {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}
					ui.renderDebug(level, pos)

					ui.renderer.Copy(ui.textureAtlas, &src, &dst)
				}
			}
		}
	}
}

func (ui *ui) drawOnFloor(level *game.Level, offsetX, offsetY int) {
	for y, row := range level.Level {
		for x, tile := range row {
			if tile.HasFloor {
				srcs := ui.textureIndex[game.BloodStained]
				src := srcs[ui.r.Intn(len(srcs))]
				if tile.BloodStained {
					if tile.Seen || tile.Visible {
						dst := sdl.Rect{X: int32(x*32 + offsetX), Y: int32(y*32 + offsetY), W: 32, H: 32}
						if tile.Seen && !tile.Visible {
							ui.textureAtlas.SetColorMod(128, 128, 128)
						} else if tile.Visible {
							ui.textureAtlas.SetColorMod(255, 255, 255)
						}

						ui.renderer.Copy(ui.textureAtlas, &src, &dst)
					}
				}
				for pos, item := range level.Items {
					tile := level.TileAtPos(pos)
					ItemSrcRect := ui.textureIndex[item.Rune][0]
					if tile.Visible || tile.Seen {
						ui.textureAtlas.SetColorMod(255, 255, 255)
						if tile.Seen && !tile.Visible {
							ui.textureAtlas.SetColorMod(128, 128, 128)
						}
						ui.renderer.Copy(ui.textureAtlas, &ItemSrcRect, &sdl.Rect{X: int32(pos.X*32 + game.OffsetX), Y: int32(pos.Y*32 + game.OffsetY), W: 32, H: 32})
					}
				}
			}
		}
	}
}

func (ui *ui) drawUI(level *game.Level) {
	eventStart := int32(float64(ui.winHeight) * .72)
	eventWidth := int32(float64(ui.winWidth) * .30)

	ui.renderer.Copy(ui.panelBackground, nil, &sdl.Rect{0, eventStart, eventWidth, int32(ui.winHeight) - eventStart})

	// loop and print events
	i := level.EventPos
	count := 0
	_, fontSizeY, _ := ui.fontSmall.SizeUTF8("A")
	for {
		event := level.Events[i]

		if event != "" {
			tex := ui.stringToTexture(event, sdl.Color{255, 0, 0, 0}, FontSmall)
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			ui.renderer.Copy(tex, nil, &sdl.Rect{5, int32(count*fontSizeY) + eventStart, w, h})
		}
		i = (i + 1) % (len(level.Events))
		count++

		if i == level.EventPos {
			break
		}
	}

	statsStart := int32(float64(ui.winHeight) * .02)
	statsHeight := int32(float64(ui.winHeight) * .50)
	statsWidth := int32(float64(ui.winWidth) * .20)

	ui.renderer.Copy(ui.panelBackground, nil, &sdl.Rect{0, statsStart, statsWidth, statsHeight})

	stats := level.Player.GetStatStrings()

	_, fontSizeY, _ = ui.fontMedium.SizeUTF8("A")
	for i, stat := range stats {
		tex := ui.stringToTexture(stat, sdl.Color{255, 255, 255, 0}, FontMedium)
		_, _, w, h, err := tex.Query()
		if err != nil {
			panic(err)
		}
		ui.renderer.Copy(tex, nil, &sdl.Rect{10, int32(i*fontSizeY) + statsStart, w, h})
	}
}

func (ui *ui) GetSinglePixelTex(colour sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = colour.R
	pixels[1] = colour.G
	pixels[2] = colour.B
	pixels[3] = colour.A
	tex.Update(nil, pixels, 4)

	return tex
}

func (ui *ui) GetInput() {
	var newLevel *game.Level
	ui.prevMouseState = getMouseState()

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Type: game.QuitGame}
				ui.QuitSDL()
				return
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Type: game.QuitGame}
					ui.QuitSDL()
					return
				}
			}
			ui.currentMouseState = getMouseState()

			var ok bool
			select {
			case newLevel, ok = <-ui.levelChan:
				if ok {
				}
			default:
			}
		}

		ui.Draw(newLevel)

		var input game.Input
		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			if ui.keyDownOnce(sdl.SCANCODE_UP) || ui.keyDownOnce(sdl.SCANCODE_W) {
				input.Type = game.Up
			} else if ui.keyDownOnce(sdl.SCANCODE_DOWN) || ui.keyDownOnce(sdl.SCANCODE_S) {
				input.Type = game.Down
			} else if ui.keyDownOnce(sdl.SCANCODE_LEFT) || ui.keyDownOnce(sdl.SCANCODE_A) {
				input.Type = game.Left
			} else if ui.keyDownOnce(sdl.SCANCODE_RIGHT) || ui.keyDownOnce(sdl.SCANCODE_D) {
				input.Type = game.Right
			}
		}

		for i, v := range ui.keyboardState {
			ui.prevKeyboardState[i] = v
		}

		if input.Type != game.None {
			ui.inputChan <- &input
		}
		ui.prevMouseState = ui.currentMouseState
		sdl.Delay(10)
	}
}
