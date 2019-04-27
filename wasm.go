package main

import (
	"math/rand"
	"strconv"
	"syscall/js"
)

var (
	ctx            js.Value
	height, width  int
	startX, startY int

	debug = 1 // 1 = show debug message, 0 = don't
)

func main() {
}

//go:export clearCanvas
func clearCanvas() {
	// Initialise doc
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
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