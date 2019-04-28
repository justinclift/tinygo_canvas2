package main

import (
	"math"
	"math/rand"
	"strconv"
	"syscall/js"
)

type matrix []float64

type Point struct {
	Label      string
	LabelAlign string
	X          float64
	Y          float64
	Z          float64
}

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

// Multiplies one matrix by another
func matrixMult(opMatrix matrix, m matrix) (resultMatrix matrix) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]
	bot0 := m[12]
	bot1 := m[13]
	bot2 := m[14]
	bot3 := m[15]

	resultMatrix = matrix{
		(opMatrix[0] * top0) + (opMatrix[1] * upperMid0) + (opMatrix[2] * lowerMid0) + (opMatrix[3] * bot0), // 1st col, top
		(opMatrix[0] * top1) + (opMatrix[1] * upperMid1) + (opMatrix[2] * lowerMid1) + (opMatrix[3] * bot1), // 2nd col, top
		(opMatrix[0] * top2) + (opMatrix[1] * upperMid2) + (opMatrix[2] * lowerMid2) + (opMatrix[3] * bot2), // 3rd col, top
		(opMatrix[0] * top3) + (opMatrix[1] * upperMid3) + (opMatrix[2] * lowerMid3) + (opMatrix[3] * bot3), // 4th col, top

		(opMatrix[4] * top0) + (opMatrix[5] * upperMid0) + (opMatrix[6] * lowerMid0) + (opMatrix[7] * bot0), // 1st col, upper middle
		(opMatrix[4] * top1) + (opMatrix[5] * upperMid1) + (opMatrix[6] * lowerMid1) + (opMatrix[7] * bot1), // 2nd col, upper middle
		(opMatrix[4] * top2) + (opMatrix[5] * upperMid2) + (opMatrix[6] * lowerMid2) + (opMatrix[7] * bot2), // 3rd col, upper middle
		(opMatrix[4] * top3) + (opMatrix[5] * upperMid3) + (opMatrix[6] * lowerMid3) + (opMatrix[7] * bot3), // 4th col, upper middle

		(opMatrix[8] * top0) + (opMatrix[9] * upperMid0) + (opMatrix[10] * lowerMid0) + (opMatrix[11] * bot0), // 1st col, lower middle
		(opMatrix[8] * top1) + (opMatrix[9] * upperMid1) + (opMatrix[10] * lowerMid1) + (opMatrix[11] * bot1), // 2nd col, lower middle
		(opMatrix[8] * top2) + (opMatrix[9] * upperMid2) + (opMatrix[10] * lowerMid2) + (opMatrix[11] * bot2), // 3rd col, lower middle
		(opMatrix[8] * top3) + (opMatrix[9] * upperMid3) + (opMatrix[10] * lowerMid3) + (opMatrix[11] * bot3), // 4th col, lower middle

		(opMatrix[12] * top0) + (opMatrix[13] * upperMid0) + (opMatrix[14] * lowerMid0) + (opMatrix[15] * bot0), // 1st col, bottom
		(opMatrix[12] * top1) + (opMatrix[13] * upperMid1) + (opMatrix[14] * lowerMid1) + (opMatrix[15] * bot1), // 2nd col, bottom
		(opMatrix[12] * top2) + (opMatrix[13] * upperMid2) + (opMatrix[14] * lowerMid2) + (opMatrix[15] * bot2), // 3rd col, bottom
		(opMatrix[12] * top3) + (opMatrix[13] * upperMid3) + (opMatrix[14] * lowerMid3) + (opMatrix[15] * bot3), // 4th col, bottom
	}
	return resultMatrix
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

// Rotates a transformation matrix around the X axis by the given degrees
func rotateAroundX(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateXMatrix := matrix{
		1, 0, 0, 0,
		0, math.Cos(rad), -math.Sin(rad), 0,
		0, math.Sin(rad), math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateXMatrix, m)
}

// Rotates a transformation matrix around the Y axis by the given degrees
func rotateAroundY(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateYMatrix := matrix{
		math.Cos(rad), 0, math.Sin(rad), 0,
		0, 1, 0, 0,
		-math.Sin(rad), 0, math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateYMatrix, m)
}

// Rotates a transformation matrix around the Z axis by the given degrees
func rotateAroundZ(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateZMatrix := matrix{
		math.Cos(rad), -math.Sin(rad), 0, 0,
		math.Sin(rad), math.Cos(rad), 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateZMatrix, m)
}

// Scales a transformation matrix by the given X, Y, and Z values
func scale(m matrix, x float64, y float64, z float64) matrix {
	scaleMatrix := matrix{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
	return matrixMult(scaleMatrix, m)
}

// Transform the XYZ co-ordinates using the values from the transformation matrix
func transform(m matrix, p Point) (t Point) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]
	//bot0 := m[12] // The fourth row values can be ignored for 3D matrices
	//bot1 := m[13]
	//bot2 := m[14]
	//bot3 := m[15]

	t.Label = p.Label
	t.LabelAlign = p.LabelAlign
	t.X = (top0 * p.X) + (top1 * p.Y) + (top2 * p.Z) + top3
	t.Y = (upperMid0 * p.X) + (upperMid1 * p.Y) + (upperMid2 * p.Z) + upperMid3
	t.Z = (lowerMid0 * p.X) + (lowerMid1 * p.Y) + (lowerMid2 * p.Z) + lowerMid3
	return
}

// Translates (moves) a transformation matrix by the given X, Y and Z values
func translate(m matrix, translateX float64, translateY float64, translateZ float64) matrix {
	translateMatrix := matrix{
		1, 0, 0, translateX,
		0, 1, 0, translateY,
		0, 0, 1, translateZ,
		0, 0, 0, 1,
	}
	return matrixMult(translateMatrix, m)
}
