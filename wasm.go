package main

import (
	"math"
	"math/rand"
	"strconv"
	"syscall/js"
)

var (
	canvasEl, ctx, doc js.Value
	graphWidth         float64
	graphHeight        float64
	height, width      int
	startX, startY     int

	debug = 1 // 1 = show debug message, 0 = don't
)

func main() {
}

//go:export clearCanvas
func clearCanvas() {
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Int()
	height = doc.Get("body").Get("clientHeight").Int()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Clear the background
	ctx.Set("fillStyle", "lightgrey")
	ctx.Call("fillRect", 0, 0, width, height)

	// Initialise starting point
	startX = rand.Intn(width)
	startY = rand.Intn(height)
}

//go:export drawLine
func drawLine() {
	endX := startX
	endY := startY

	// Generate new random start point
	startX = rand.Intn(width)
	startY = rand.Intn(height)

	// Generate random colour
	colR := rand.Intn(255)
	colG := rand.Intn(255)
	colB := rand.Intn(255)

	ctx.Set("strokeStyle", "rgb("+strconv.Itoa(colR)+", "+strconv.Itoa(colG)+", "+strconv.Itoa(colB)+")")
	ctx.Set("lineWidth", "5")
	ctx.Call("beginPath")
	ctx.Call("moveTo", startX, startY)
	ctx.Call("lineTo", endX, endY)
	ctx.Call("stroke")
}

// Simple mouse handler watching for people clicking on the source code link
//go:export mouseDownHandler
func mouseDownHandler(clientX int, clientY int) {
	if debug == 1 {
		println("ClientX: " + strconv.Itoa(clientX) + " ClientY: " + strconv.Itoa(clientY))
	}

	endX := startX
	endY := startY

	// TODO: Although this is receiving the co-ordinates ok, the lines being drawn appear scaled down to a small area.
	// TODO  Find out why and fix it.
	startX = clientX
	startY = clientY

	// Generate random colour
	colR := rand.Intn(255)
	colG := rand.Intn(255)
	colB := rand.Intn(255)

	ctx.Set("strokeStyle", "rgb("+strconv.Itoa(colR)+", "+strconv.Itoa(colG)+", "+strconv.Itoa(colB)+")")
	ctx.Set("lineWidth", "5")
	ctx.Call("beginPath")
	ctx.Call("moveTo", startX, startY)
	ctx.Call("lineTo", endX, endY)
	ctx.Call("stroke")
}

// Renders one frame of the animation
func renderFrame() {
	// Handle window resizing
	curBodyW := doc.Get("body").Get("clientWidth").Int()
	curBodyH := doc.Get("body").Get("clientHeight").Int()
	if curBodyW != width || curBodyH != height {
		width, height = curBodyW, curBodyH
		canvasEl.Set("width", width)
		canvasEl.Set("height", height)
	}

	// Setup useful variables
	border := float64(2)
	gap := float64(3)
	left := border + gap
	top := border + gap
	graphWidth = float64(width) * 0.75
	graphHeight = float64(height) - 1
	//centerX := graphWidth / 2
	//centerY := graphHeight / 2

	// Clear the background
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", 0, 0, width, height)

	// Draw grid lines
	step := math.Min(float64(width), float64(height)) / 30
	ctx.Set("strokeStyle", "rgb(220, 220, 220)")
	ctx.Call("setLineDash", []interface{}{1, 3})
	for i := left; i < graphWidth-step; i += step {
		// Vertical dashed lines
		ctx.Call("beginPath")
		ctx.Call("moveTo", i+step, top)
		ctx.Call("lineTo", i+step, graphHeight)
		ctx.Call("stroke")
	}
	for i := top; i < graphHeight-step; i += step {
		// Horizontal dashed lines
		ctx.Call("beginPath")
		ctx.Call("moveTo", left, i+step)
		ctx.Call("lineTo", graphWidth-border, i+step)
		ctx.Call("stroke")
	}
}