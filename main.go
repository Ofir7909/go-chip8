package main

import (
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowTitle  = "Go Chip-8!"
	windowWidth  = 512
	windowHeight = 256
)

var onColor = [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
var offColor = [4]byte{0xFF, 0x25, 0x36, 0x1d}

var keymap = [16]sdl.Scancode{
	sdl.SCANCODE_X,
	sdl.SCANCODE_1,
	sdl.SCANCODE_2,
	sdl.SCANCODE_3,
	sdl.SCANCODE_Q,
	sdl.SCANCODE_W,
	sdl.SCANCODE_E,
	sdl.SCANCODE_A,
	sdl.SCANCODE_S,
	sdl.SCANCODE_D,
	sdl.SCANCODE_Z,
	sdl.SCANCODE_C,
	sdl.SCANCODE_4,
	sdl.SCANCODE_R,
	sdl.SCANCODE_F,
	sdl.SCANCODE_V,
}

type sdlContext struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Missing Arguments.\nUsage: %v <path/to/rom>", os.Args[0])
	}
	var context sdlContext
	var cpu Cpu

	initialize(&context)
	cpu.reset()
	loadRom(&cpu, os.Args[1])

	quit := false
	for !quit {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch event := e.(type) {
			case *sdl.QuitEvent:
				quit = true
			case *sdl.KeyboardEvent:
				for i, v := range keymap {
					if v == event.Keysym.Scancode {
						if event.Type == sdl.KEYDOWN {
							cpu.keys[i] = 1
						} else if event.Type == sdl.KEYUP {
							cpu.keys[i] = 0
						}
					}
				}

			}
		}

		cpu.step()

		render(&cpu, &context)
		time.Sleep(16 * time.Millisecond)
	}

	shutdown(&context)
}

func initialize(context *sdlContext) {
	var err error
	var sdlFlags uint32 = sdl.INIT_EVERYTHING
	err = sdl.Init(sdlFlags)
	if err != nil {
		log.Fatalf("Could not initialize SDL2: %v", err)
	}

	var windowFlags uint32 = sdl.WINDOW_RESIZABLE | sdl.WINDOW_SHOWN
	context.window, err = sdl.CreateWindow(windowTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowWidth, windowHeight, windowFlags)
	if err != nil {
		log.Fatalf("Could not create window: %v", err)
	}

	context.renderer, err = sdl.CreateRenderer(context.window, -1, 0)
	if err != nil {
		log.Fatalf("Could not create renderer: %v", err)
	}

	context.texture, err = context.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, int32(screenWidth), int32(screenHeight))
	if err != nil {
		log.Fatalf("Could not create texture: %v", err)
	}
}

func render(cpu *Cpu, context *sdlContext) {
	textureRect := sdl.Rect{0, 0, int32(screenWidth), int32(screenHeight)}
	var bytes []byte
	var err error
	bytes, _, err = context.texture.Lock(&textureRect)
	if err != nil {
		log.Fatalf("Failed to update texture")
	}

	for i, v := range cpu.graphics {
		if v != 0 {
			copy(bytes[i*4:i*4+4], onColor[:])
		} else {
			copy(bytes[i*4:i*4+4], offColor[:])
		}
	}
	context.texture.Unlock()

	context.renderer.Clear()
	context.renderer.Copy(context.texture, &sdl.Rect{0, 0, 64, 32}, &sdl.Rect{0, 0, windowWidth, windowHeight})
	context.renderer.Present()
}

func shutdown(context *sdlContext) {
	var err error

	err = context.texture.Destroy()
	if err != nil {
		log.Fatalf("Failed to destroy texture: %v", err)
	}

	err = context.renderer.Destroy()
	if err != nil {
		log.Fatalf("Failed to destroy renderer: %v", err)
	}

	err = context.window.Destroy()
	if err != nil {
		log.Fatalf("Failed to destroy window: %v", err)
	}

	sdl.Quit()
}

func loadRom(cpu *Cpu, filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Failed to open file '%v'", filepath)
	}
	for i, v := range data {
		cpu.memory[programBaseAddress+uint16(i)] = v
	}
}
