package main

import (
	"fmt"
	"gonesem/nes"
	"gonesem/nes/cartridge"
	"gonesem/nes/color"
	"image"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const width, height, scale = 256, 240, 3
const title = "NES"
const fps float64 = (1.0 / 60.0)

const vertexShaderSource = `
	#version 460

	layout (location = 0) in vec3 inPos;
	layout (location = 1) in vec2 inTexCoord;

	out vec2 TexCoord;

	void main() {
		gl_Position = vec4(inPos, 1.0);
		TexCoord = vec2(inTexCoord.x, 1.0 - inTexCoord.y);
	}
` + "\x00"

const fragmentShaderSource = `
	#version 460

	in vec2 TexCoord;

	out vec4 fragColor;

	uniform sampler2D quadTexture;

	void main() {
		fragColor = texture(quadTexture, TexCoord);
	}
` + "\x00"

var (
	quad = []float32{
		// Top Left
		-1.0, 1.0, 0.0, // Position
		1.0, 0.0, // Texture Coordinates

		// Top Right
		1.0, 1.0, 0.0,
		0.0, 0.0,

		// Bottom Right
		1.0, -1.0, 0.0,
		0.0, 1.0,

		// Bottom Left
		-1.0, -1.0, 0.0,
		1.0, 1.0,
	}
)

var (
	indices = []uint32{
		0, 1, 2,
		0, 2, 3,
	}
)

func init() {
	runtime.LockOSThread()
}

func main() {
	nes, err := nesInit()

	if err != nil {
		log.Fatalf("Failed to initialize NES console: %s\n", err)

		os.Exit(1)
	}

	window, err := glfwInit()

	if err != nil {
		log.Fatalf("Failed to initialize GLFW: %s\n", err)

		os.Exit(1)
	}

	defer glfw.Terminate()

	program, err := glInit()

	if err != nil {
		log.Fatalf("Failed to initialize OpenGL: %s\n", err)

		os.Exit(1)
	}

	gl.UseProgram(program)

	vao := createVao(quad, indices)
	texture := createTexture()

	gl.ClearColor(0, 0, 0, 1)

	timestamp := glfw.GetTime()
	residualTime := 0.0

	for !window.ShouldClose() {
		deltaTime := glfw.GetTime() - timestamp

		if residualTime > 0 {
			residualTime -= deltaTime
		} else {
			residualTime += fps - deltaTime

			gl.Clear(gl.COLOR_BUFFER_BIT)

			nes.NextFrame()

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, texture)

			setFrameTexture(nes.GetFrame())

			gl.BindVertexArray(vao)
			gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, unsafe.Pointer(nil))
			gl.BindVertexArray(0)

			window.SwapBuffers()
			glfw.PollEvents()
		}

		timestamp = glfw.GetTime()
	}
}

func nesInit() (*nes.NES, error) {
	cartridge, err := cartridge.NewCartridge("../test/data/roms/Donkey Kong.nes")

	if err != nil {
		return nil, err
	}

	colorPalette, err := color.NewColorPalette("../test/data/pals/NESdev.pal")

	if err != nil {
		return nil, err
	}

	nes := nes.NewNES(cartridge, colorPalette)

	return nes, nil
}

func glfwInit() (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	window, err := glfw.CreateWindow(width*scale, height*scale, title, nil, nil)

	if err != nil {
		return nil, err
	}

	window.MakeContextCurrent()

	return window, nil
}

func glInit() (uint32, error) {
	if err := gl.Init(); err != nil {
		return 0, err
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL Version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)

	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)

	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)

	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile: %v", log)
	}

	return program, nil
}

func createVao(vertices []float32, indices []uint32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)

	var ebo uint32
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	// Copy vertices data to vertex buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Copy indices to element buffer
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Position
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 5*4, uintptr(0))
	gl.EnableVertexAttribArray(0)

	// Texture position
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, uintptr(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	glslSrc, freeFn := gl.Strs(source)
	gl.ShaderSource(shader, 1, glslSrc, nil)
	freeFn()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)

	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func createTexture() uint32 {
	var texture uint32

	gl.GenTextures(1, &texture)

	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return texture
}

func setFrameTexture(image *image.RGBA) {
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(image.Rect.Size().X),
		int32(image.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(image.Pix))
}
