package ui

import (
	"log"
	"runtime"
	"fmt"
	
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"github.com/BrianWill/nes/nes"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/gordonklaus/portaudio"
)

const (
	width  = 256
	height = 240
	scale  = 3
	title  = "NES"
)

const padding = 0

type GameView struct {
	director *Director
	console  *nes.Console
	title    string
	hash     string
	texture  uint32
	record   bool
	frames   []image.Image
}

var fontMask image.Image

func DrawCenteredText(dst draw.Image, text string, dx, dy int, c color.Color) {
	// split text into rows
	var rows []string
	const rowMaxLength = 15
	words := strings.Fields(text)
	if len(words) == 0 {
		row := words[0]
		for _, word := range words[1:] {
			newRow := row + " " + word
			if len(newRow) <= rowMaxLength {
				row = newRow
			} else {
				rows = append(rows, row)
				row = word
			}
		}
		rows = append(rows, row)
	}
	// draw all rows
	for i, rowText := range rows {
		x := 128 - len(rowText)*8
		y := 120 - len(rows)*12 + i*24
		x += dx
		y += dy
		// draw row
		for i := range text {
			// draw character
			ch := rowText[i]
			if !(ch < 32 || ch > 128) {
				cx := int((ch-32)%16) * 16
				cy := int((ch-32)/16) * 16
				r := image.Rect(x, y, x+16, y+16)
				src := &image.Uniform{c}
				sp := image.Pt(cx, cy)
				draw.DrawMask(dst, r, src, sp, fontMask, sp, draw.Over)	
			}
			x += 16  // next char pos
		}
	}
}


var fontData = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x60,
	0x02, 0x03, 0x00, 0x00, 0x00, 0x8F, 0x9F, 0x44, 0x1B, 0x00, 0x00, 0x00,
	0x06, 0x50, 0x4C, 0x54, 0x45, 0x00, 0x00, 0x01, 0xFF, 0xFF, 0xFE, 0x6A,
	0x62, 0xC8, 0x2E, 0x00, 0x00, 0x00, 0x01, 0x62, 0x4B, 0x47, 0x44, 0x00,
	0x88, 0x05, 0x1D, 0x48, 0x00, 0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73,
	0x00, 0x00, 0x0E, 0xC4, 0x00, 0x00, 0x0E, 0xC4, 0x01, 0x95, 0x2B, 0x0E,
	0x1B, 0x00, 0x00, 0x02, 0xD2, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0xC5,
	0x98, 0x49, 0x62, 0x23, 0x21, 0x0C, 0x45, 0xB5, 0xE1, 0x7E, 0xDA, 0xFC,
	0xFB, 0x5F, 0xA5, 0x4B, 0xE8, 0x4B, 0x88, 0x72, 0x2A, 0x1D, 0x07, 0xB9,
	0x9B, 0xC4, 0x2E, 0x8C, 0xE1, 0x19, 0x8D, 0x0C, 0x22, 0x56, 0x86, 0x5E,
	0xFF, 0x98, 0x2F, 0x81, 0x78, 0x81, 0xB5, 0xC5, 0x77, 0xF2, 0x50, 0xE2,
	0x9B, 0x63, 0x00, 0x90, 0x80, 0x71, 0xD5, 0xAD, 0xE6, 0x9F, 0xC5, 0x81,
	0x90, 0xD9, 0x92, 0xF0, 0x2C, 0xF9, 0x6B, 0x4D, 0x00, 0x00, 0xEA, 0xFD,
	0x31, 0x45, 0x40, 0x4C, 0xD2, 0xDE, 0x50, 0xC4, 0xBB, 0x4B, 0xD0, 0x02,
	0xF0, 0xCF, 0x26, 0x82, 0xBA, 0xE2, 0xE0, 0xDF, 0x06, 0x00, 0x5E, 0xD8,
	0x1D, 0xBB, 0x04, 0xE7, 0x00, 0x4E, 0xC5, 0x44, 0x98, 0xAD, 0xF6, 0x3F,
	0x20, 0xF2, 0x17, 0x11, 0x96, 0x79, 0xBB, 0x00, 0xAE, 0x48, 0xA5, 0x11,
	0xCB, 0x34, 0xAB, 0x19, 0xB1, 0x2C, 0x08, 0x69, 0x03, 0x24, 0x87, 0xD3,
	0xB7, 0xEE, 0x48, 0xEA, 0xEE, 0xCA, 0x05, 0xA0, 0x7D, 0x80, 0xE7, 0x68,
	0x79, 0x0C, 0x9F, 0x9F, 0xB4, 0xFE, 0x10, 0x90, 0x8E, 0x33, 0x03, 0xC9,
	0x2A, 0xEE, 0x28, 0x70, 0xD7, 0x66, 0xBB, 0xBC, 0x2A, 0x5C, 0xDC, 0xE9,
	0x8E, 0x01, 0x23, 0x03, 0x67, 0xD0, 0x89, 0x90, 0x41, 0x54, 0x82, 0x7C,
	0x99, 0x70, 0x33, 0x23, 0x1A, 0x00, 0x48, 0x3B, 0x79, 0x0E, 0x63, 0x58,
	0xB3, 0x1F, 0xE4, 0x0E, 0x70, 0xAF, 0x66, 0xA5, 0x17, 0x10, 0x8A, 0xA3,
	0x00, 0xD1, 0x29, 0x95, 0x7B, 0xF3, 0xA3, 0x45, 0xEB, 0x02, 0xB8, 0x4D,
	0xE7, 0x40, 0x95, 0x32, 0x6D, 0x57, 0xEE, 0x4B, 0xE0, 0xC6, 0xE0, 0x06,
	0xC0, 0x15, 0x36, 0x9E, 0xB2, 0x62, 0x36, 0x37, 0xC5, 0x2D, 0xF3, 0x2E,
	0x33, 0xA2, 0x26, 0x97, 0x63, 0xC0, 0xB4, 0x9E, 0xA5, 0x73, 0xBA, 0xED,
	0x1C, 0x12, 0xEE, 0x6B, 0x7A, 0x02, 0x13, 0x6D, 0x00, 0x8A, 0x2B, 0x3B,
	0xE4, 0x18, 0xF0, 0x46, 0x79, 0x23, 0xEA, 0xDF, 0x00, 0xF8, 0x54, 0xB5,
	0x84, 0xB0, 0xFA, 0x12, 0xE7, 0x15, 0xF5, 0xF4, 0x36, 0xE5, 0xF4, 0x7A,
	0x09, 0xB2, 0x39, 0xF6, 0x18, 0x80, 0x4C, 0xE2, 0xF6, 0xF4, 0x5A, 0xD5,
	0xD7, 0x0A, 0xB6, 0xE8, 0x65, 0xBF, 0x67, 0x7F, 0x69, 0xEA, 0x53, 0x00,
	0xA7, 0xC6, 0x0D, 0x8E, 0xAC, 0xE7, 0xEE, 0xEA, 0x13, 0x90, 0x71, 0x67,
	0x22, 0xB6, 0x03, 0x80, 0x1C, 0x00, 0x44, 0x52, 0xA1, 0x32, 0x03, 0xB0,
	0x12, 0x8A, 0x82, 0x0A, 0x6E, 0x02, 0x00, 0x0F, 0x22, 0x70, 0x9A, 0x54,
	0x76, 0xB0, 0x30, 0x4C, 0x9C, 0x36, 0x40, 0xE9, 0xF8, 0x6A, 0xC6, 0x1B,
	0x60, 0xA0, 0xA4, 0xFB, 0x2E, 0xC0, 0x80, 0xA6, 0xE2, 0x36, 0x57, 0xF6,
	0x3D, 0x4F, 0xBA, 0xF2, 0xA0, 0x52, 0x07, 0x96, 0xD9, 0x19, 0x0B, 0x87,
	0x80, 0xFF, 0x5F, 0xC0, 0xF4, 0x05, 0xA6, 0xAE, 0x3A, 0xC5, 0x4D, 0xB9,
	0x53, 0x0C, 0x6E, 0x7F, 0x7D, 0x59, 0xEF, 0x02, 0x6C, 0x9D, 0xDD, 0x57,
	0xB7, 0xB6, 0x19, 0xB8, 0x1E, 0xC4, 0xB9, 0xC4, 0xF8, 0xC7, 0x8F, 0x00,
	0x64, 0x07, 0x0C, 0x37, 0xD5, 0x04, 0x8C, 0x72, 0x00, 0xE1, 0x36, 0xA0,
	0x1B, 0xC0, 0x85, 0x25, 0x00, 0x1A, 0x89, 0x35, 0x96, 0xFE, 0x75, 0x08,
	0x8B, 0x3E, 0x4D, 0x00, 0x33, 0xE1, 0x0C, 0xE9, 0x75, 0xC6, 0x0B, 0x37,
	0xCD, 0xE5, 0xCE, 0x25, 0xE0, 0x10, 0xEC, 0x2B, 0xCC, 0x39, 0x80, 0x5B,
	0x29, 0x0B, 0xD5, 0x1A, 0xBA, 0x73, 0xB1, 0x71, 0x41, 0x34, 0x96, 0xFF,
	0x3C, 0x04, 0xB5, 0x03, 0x2E, 0x09, 0x36, 0x25, 0xFA, 0xF3, 0x96, 0xD2,
	0xE8, 0xCA, 0x57, 0x89, 0x67, 0x13, 0xE0, 0x30, 0x92, 0xD1, 0x00, 0xC0,
	0xC3, 0x11, 0x8E, 0xD1, 0x52, 0xF5, 0x95, 0x8E, 0xF4, 0xD5, 0x01, 0xBC,
	0x1F, 0x50, 0x1F, 0x7C, 0xFB, 0x10, 0x80, 0x1B, 0x6D, 0xDF, 0xAE, 0xF8,
	0xB6, 0x46, 0xD6, 0x2A, 0xAA, 0x37, 0x11, 0x34, 0x0E, 0xE6, 0xEC, 0xDF,
	0x01, 0xE0, 0xE4, 0x06, 0x13, 0x57, 0xD6, 0x81, 0x78, 0x16, 0x11, 0x06,
	0x3D, 0xD9, 0xF6, 0x1A, 0x31, 0xE2, 0x23, 0x00, 0xC4, 0xF1, 0x87, 0xAD,
	0x09, 0xD0, 0x7F, 0x05, 0x58, 0x97, 0x0E, 0x3B, 0x20, 0x82, 0xAD, 0x1D,
	0x60, 0x99, 0xE3, 0x66, 0xC6, 0x32, 0xED, 0xF1, 0x22, 0x4E, 0x00, 0xD2,
	0x8C, 0xE7, 0x80, 0xE7, 0x78, 0xFD, 0xBA, 0xFE, 0x46, 0xC0, 0xFF, 0x02,
	0xF0, 0x5D, 0x3F, 0xE6, 0xF6, 0x99, 0x60, 0xF1, 0x74, 0x7E, 0xF9, 0x28,
	0x60, 0x9D, 0x01, 0xFC, 0xA2, 0x6A, 0x74, 0x00, 0xC2, 0x7C, 0xB6, 0x9C,
	0x4A, 0x6C, 0x75, 0x96, 0x63, 0xC1, 0x5F, 0x74, 0xEF, 0x02, 0x58, 0xF7,
	0x2F, 0xC7, 0x80, 0x0C, 0x5F, 0x8A, 0x84, 0xEA, 0xDA, 0xB6, 0xDE, 0x84,
	0xCB, 0x22, 0x77, 0x35, 0x32, 0xD6, 0xF5, 0x61, 0x23, 0x40, 0x25, 0xAF,
	0xFC, 0x36, 0x80, 0x75, 0x56, 0xD4, 0xE3, 0xBE, 0x48, 0xD9, 0xDD, 0x34,
	0x00, 0x52, 0x89, 0xB9, 0xC3, 0xF2, 0x04, 0x7E, 0xA5, 0xAF, 0x10, 0xA1,
	0xDC, 0x56, 0x7C, 0x02, 0xA0, 0x92, 0x8B, 0x06, 0xF7, 0xFF, 0xB2, 0xD2,
	0x3A, 0xAF, 0x21, 0xFD, 0x88, 0xAC, 0xD5, 0x95, 0xF3, 0x0A, 0xB5, 0x0D,
	0xF0, 0xEB, 0x7B, 0x83, 0x53, 0xC0, 0x1F, 0xEF, 0x0D, 0xA2, 0x4D, 0x77,
	0x69, 0xB8, 0xB7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
	0x42, 0x60, 0x82,
}


func (view *GameView) onKey(window *glfw.Window,
	key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeySpace:
			screenshot(view.console.Buffer())
		case glfw.KeyR:
			view.console.Reset()
		case glfw.KeyTab:
			if view.record {
				view.record = false
				animation(view.frames)
				view.frames = nil
			} else {
				view.record = true
			}
		}
	}
}

func init() {
	// we need a parallel OS thread to avoid audio stuttering
	runtime.GOMAXPROCS(2)

	// we need to keep OpenGL calls on a single thread
	runtime.LockOSThread()
}

func Run(paths []string) {
	fmt.Println("start run")
	// initialize audio
	portaudio.Initialize()
	defer portaudio.Terminate()

	audio := NewAudio()
	if err := audio.Start(); err != nil {
		log.Fatalln(err)
	}
	defer audio.Stop()

	// initialize fontMask
	{
		im, err := png.Decode(bytes.NewBuffer(fontData))
		if err != nil {
			log.Fatalln(err)
		}
		size := im.Bounds().Size()
		mask := image.NewRGBA(im.Bounds())
		for y := 0; y < size.Y; y++ {
			for x := 0; x < size.X; x++ {
				r, _, _, _ := im.At(x, y).RGBA()
				if r > 0 {
					mask.Set(x, y, color.Opaque)
				}
			}
		}
		fontMask = mask
	}

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	// create window
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(width*scale, height*scale, title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}
	window.MakeContextCurrent()

	// initialize gl
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	gl.Enable(gl.TEXTURE_2D)

	// run director
	d := &Director{window: window, audio: audio}
	d.menuView = &MenuView{director: d, paths: paths, texture: NewTexture()}

	if len(paths) == 1 {
		d.PlayGame(paths[0])
	} else {
		d.SetView(d.menuView)
	}

	// main loop
	for !d.window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		timestamp := glfw.GetTime()
		dt := timestamp - d.timestamp
		d.timestamp = timestamp
		if d.view != nil {
			// update view
			switch v := d.view.(type) {
			case *GameView:
				if dt > 1 {
					dt = 0
				}
				window := v.director.window
				console := v.console
				if joystickReset(glfw.Joystick1) || joystickReset(glfw.Joystick2) || readKey(window, glfw.KeyEscape) {
					d.SetView(d.menuView)
				}
				// update controllers
				{
					turbo := console.PPU.Frame%6 < 3
					k1 := readKeys(window, turbo)
					j1 := readJoystick(glfw.Joystick1, turbo)
					j2 := readJoystick(glfw.Joystick2, turbo)
					console.SetButtons1(combineButtons(k1, j1))
					console.SetButtons2(j2)
				}
				console.StepSeconds(dt)
				gl.BindTexture(gl.TEXTURE_2D, v.texture)
				setTexture(console.Buffer())
				// draw buffer
				{
					window := v.director.window
					w, h := window.GetFramebufferSize()
					s1 := float32(w) / 256
					s2 := float32(h) / 240
					f := float32(1 - padding)
					var x, y float32
					if s1 >= s2 {
						x = f * s2 / s1
						y = f
					} else {
						x = f
						y = f * s1 / s2
					}
					gl.Begin(gl.QUADS)
					gl.TexCoord2f(0, 1)
					gl.Vertex2f(-x, -y)
					gl.TexCoord2f(1, 1)
					gl.Vertex2f(x, -y)
					gl.TexCoord2f(1, 0)
					gl.Vertex2f(x, y)
					gl.TexCoord2f(0, 0)
					gl.Vertex2f(-x, y)
					gl.End()
				}

				gl.BindTexture(gl.TEXTURE_2D, 0)
				if v.record {
					v.frames = append(v.frames, copyImage(console.Buffer()))
				}
			case *MenuView:
				// check buttons
				{
					window := v.director.window
					k1 := readKeys(window, false)
					j1 := readJoystick(glfw.Joystick1, false)
					j2 := readJoystick(glfw.Joystick2, false)
					buttons := combineButtons(combineButtons(j1, j2), k1)
					now := glfw.GetTime()
					for i := range buttons {
						if buttons[i] && !v.buttons[i] {
							v.times[i] = now + initialDelay
							v.onPress(i)
						} else if !buttons[i] && v.buttons[i] {
							v.onRelease(i)
						} else if buttons[i] && now >= v.times[i] {
							v.times[i] = now + repeatDelay
							v.onPress(i)
						}
					}
					v.buttons = buttons
				}
				// purge texture
				for {
					select {
					case path := <-v.texture.ch:
						delete(v.texture.lookup, path)
					default:
						return
					}
				}
				window := v.director.window
				w, h := window.GetFramebufferSize()
				sx := 256 + margin*2
				sy := 240 + margin*2
				nx := (w - border*2) / sx
				ny := (h - border*2) / sy
				ox := (w-nx*sx)/2 + margin
				oy := (h-ny*sy)/2 + margin
				if nx < 1 {
					nx = 1
				}
				if ny < 1 {
					ny = 1
				}
				v.nx = nx
				v.ny = ny

				// clamp selection
				{
					if v.i < 0 {
						v.i = v.nx - 1
					}
					if v.i >= v.nx {
						v.i = 0
					}
					if v.j < 0 {
						v.j = 0
						v.scroll--
					}
					if v.j >= v.ny {
						v.j = v.ny - 1
						v.scroll++
					}
					v.clampScroll(true)
				}

				gl.PushMatrix()
				gl.Ortho(0, float64(w), float64(h), 0, -1, 1)
				gl.BindTexture(gl.TEXTURE_2D, v.texture.texture)
				for j := 0; j < ny; j++ {
					for i := 0; i < nx; i++ {
						x := float32(ox + i*sx)
						y := float32(oy + j*sy)
						index := nx*(j+v.scroll) + i
						if index >= len(v.paths) {
							continue
						}
						path := v.paths[index]

						if idx, ok := v.texture.lookup[path]; ok {
							index = idx
						} else {
							index = loadTexture(v.texture, path)
						}
						// texture coords
						tx := float32(index % textureDim) / textureDim
						ty := float32(index / textureDim) / textureDim
						tw := float32(1.0) / textureDim
						th := tw * 240 / 256
						
						// draw thumbnail
						sx := x + 4
						sy := y + 4
						gl.Disable(gl.TEXTURE_2D)
						gl.Color3f(0.2, 0.2, 0.2)
						gl.Begin(gl.QUADS)
						gl.Vertex2f(sx, sy)
						gl.Vertex2f(sx+256, sy)
						gl.Vertex2f(sx+256, sy+240)
						gl.Vertex2f(sx, sy+240)
						gl.End()
						gl.Enable(gl.TEXTURE_2D)
						gl.Color3f(1, 1, 1)
						gl.Begin(gl.QUADS)
						gl.TexCoord2f(tx, ty)
						gl.Vertex2f(x, y)
						gl.TexCoord2f(tx+tw, ty)
						gl.Vertex2f(x+256, y)
						gl.TexCoord2f(tx+tw, ty+th)
						gl.Vertex2f(x+256, y+240)
						gl.TexCoord2f(tx, ty+th)
						gl.Vertex2f(x, y+240)
						gl.End()
					}
				}
				gl.BindTexture(gl.TEXTURE_2D, 0)  // unbind
				if int((timestamp - v.t)*4)%2 == 0 {
					x := float32(ox + v.i*sx)
					y := float32(oy + v.j*sy)
					var p, w float32 = 8, 4

					// draw selection highlight border
					gl.LineWidth(w)
					gl.Begin(gl.LINE_STRIP)
					gl.Vertex2f(x-p, y-p)
					gl.Vertex2f(x+256+p, y-p)
					gl.Vertex2f(x+256+p, y+240+p)
					gl.Vertex2f(x-p, y+240+p)
					gl.Vertex2f(x-p, y-p)
					gl.End()
				}
				gl.PopMatrix()
			}
		}
		d.window.SwapBuffers()
		glfw.PollEvents()
	}
	d.SetView(nil)
}
