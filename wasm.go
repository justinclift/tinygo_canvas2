package main

import (
	"fmt"
	"syscall/js"
)

func main() {
}

//go:export clearCanvas
func clearCanvas() {

	fmt.Println("Clearing canvas...")

	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width := doc.Get("body").Get("clientWidth").Float()
	height := doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx := canvasEl.Call("getContext", "2d")

	// Clear the background
	ctx.Set("fillStyle", "red")
	ctx.Call("fillRect", 0, 0, width, height)

	fmt.Println("Canvas cleared")

	// Draw a line
	ctx.Set("strokeStyle", "blue")
	ctx.Set("lineWidth", "5")
	//ctx.Call("setLineDash", []interface{}{1, 3})
	ctx.Call("beginPath")
	ctx.Call("moveTo", 0, 0)
	ctx.Call("lineTo", width, height)
	ctx.Call("stroke")
}
