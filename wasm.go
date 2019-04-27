package main

import (
	"math/rand"
	"strconv"
	"syscall/js"
)

var (
	ctx           js.Value
	height, width int
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
}

//go:export drawLine
func drawLine() {
	// Generate random start and end points
	startX := rand.Intn(width)
	startY := rand.Intn(height)
	endX := rand.Intn(width)
	endY := rand.Intn(height)

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
