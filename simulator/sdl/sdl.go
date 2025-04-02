package sdl

import (
	"github.com/crookdc/nand2tetris/simulator"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	keymap = map[sdl.Keycode]uint16{
		sdl.K_SPACE:          32,
		sdl.K_EXCLAIM:        33,
		sdl.K_QUOTEDBL:       34,
		sdl.K_HASH:           35,
		sdl.K_DOLLAR:         36,
		sdl.K_PERCENT:        37,
		sdl.K_AMPERSAND:      38,
		sdl.K_QUOTE:          39,
		sdl.K_LEFTPAREN:      40,
		sdl.K_RIGHTPAREN:     41,
		sdl.K_ASTERISK:       42,
		sdl.K_PLUS:           43,
		sdl.K_COMMA:          44,
		sdl.K_MINUS:          45,
		sdl.K_PERIOD:         46,
		sdl.K_SLASH:          47,
		sdl.K_0:              48,
		sdl.K_1:              49,
		sdl.K_2:              50,
		sdl.K_3:              51,
		sdl.K_4:              52,
		sdl.K_5:              53,
		sdl.K_6:              54,
		sdl.K_7:              55,
		sdl.K_8:              56,
		sdl.K_9:              57,
		sdl.K_COLON:          58,
		sdl.K_SEMICOLON:      59,
		sdl.K_LESS:           60,
		sdl.K_EQUALS:         61,
		sdl.K_GREATER:        62,
		sdl.K_QUESTION:       63,
		sdl.K_AT:             64,
		sdl.K_LEFTBRACKET:    91,
		sdl.K_RIGHTBRACKET:   93,
		sdl.K_CARET:          94,
		sdl.K_UNDERSCORE:     95,
		sdl.K_BACKQUOTE:      96,
		sdl.K_a:              97,
		sdl.K_b:              98,
		sdl.K_c:              99,
		sdl.K_d:              100,
		sdl.K_e:              101,
		sdl.K_f:              102,
		sdl.K_g:              103,
		sdl.K_h:              104,
		sdl.K_i:              105,
		sdl.K_j:              106,
		sdl.K_k:              107,
		sdl.K_l:              108,
		sdl.K_m:              109,
		sdl.K_n:              110,
		sdl.K_o:              111,
		sdl.K_p:              112,
		sdl.K_q:              113,
		sdl.K_r:              114,
		sdl.K_s:              115,
		sdl.K_t:              116,
		sdl.K_u:              117,
		sdl.K_v:              118,
		sdl.K_w:              119,
		sdl.K_x:              120,
		sdl.K_y:              121,
		sdl.K_z:              122,
		sdl.K_KP_LEFTBRACE:   123,
		sdl.K_KP_VERTICALBAR: 124,
		sdl.K_KP_RIGHTBRACE:  125,
		sdl.K_DELETE:         127,
		sdl.K_KP_ENTER:       128,
		sdl.K_BACKSPACE:      129,
		sdl.K_LEFT:           130,
		sdl.K_UP:             131,
		sdl.K_RIGHT:          132,
		sdl.K_DOWN:           133,
		sdl.K_ESCAPE:         140,
	}
)

func NewScreen() (*Screen, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}
	window, err := sdl.CreateWindow("Hack", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 512, 256, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, 0, 0)
	if err != nil {
		return nil, err
	}
	screen := Screen{
		window:   window,
		renderer: renderer,
	}
	if err := screen.Clear(); err != nil {
		return nil, err
	}
	screen.renderer.Present()
	return &screen, nil
}

type Screen struct {
	window   *sdl.Window
	renderer *sdl.Renderer
}

func (s *Screen) Clear() error {
	if err := s.renderer.SetDrawColor(0, 0, 0, 255); err != nil {
		return err
	}
	if err := s.renderer.Clear(); err != nil {
		return err
	}
	return nil
}

func (s *Screen) Fill(color simulator.Color, points ...simulator.Point) error {
	if len(points) == 0 {
		return nil
	}
	if err := s.renderer.SetDrawColor(color.R, color.G, color.B, 255); err != nil {
		return err
	}
	converted := make([]sdl.Point, len(points))
	for i, point := range points {
		converted[i] = sdl.Point{
			X: int32(point.X),
			Y: int32(point.Y),
		}
	}
	return s.renderer.DrawPoints(converted)
}

func (s *Screen) Present() {
	s.renderer.Present()
}

func (s *Screen) Close() {
	_ = s.window.Destroy()
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		capitalize: false,
	}
}

type Keyboard struct {
	capitalize bool
	current    uint16
}

func (k *Keyboard) Poll() uint16 {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.KeyboardEvent:
			if e.State == sdl.RELEASED {
				k.current = 0
				continue
			} else if c, ok := keymap[e.Keysym.Sym]; ok {
				k.current = c
			}
		}
	}
	return k.current
}
