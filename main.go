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
			if e.GetType() == sdl.QUIT {
				quit = true
			}
		}

		cpu.step()

		render(&context)
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

	context.texture, err = context.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, 64, 32)
	if err != nil {
		log.Fatalf("Could not create texture: %v", err)
	}
}

func render(context *sdlContext) {
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
