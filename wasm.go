package main

import (
	"math"
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

type Edge []int
type Surface []int

type Object struct {
	C         string // Colour of the object
	P         []Point
	E         []Edge    // List of points to connect by edges
	S         []Surface // List of points to connect in order, to create a surface
	DrawOrder int       // Draw order for the object
	Name      string
}

const (
	KEY_MOVE_LEFT int = iota + 1
	KEY_MOVE_RIGHT
	KEY_MOVE_UP
	KEY_MOVE_DOWN
	KEY_ROTATE_LEFT
	KEY_ROTATE_RIGHT
	KEY_ROTATE_UP
	KEY_ROTATE_DOWN
	KEY_PAGE_UP
	KEY_PAGE_DOWN
	KEY_HOME
	KEY_END
	KEY_MINUS
	KEY_PLUS
)

type OperationType int

const (
	NOTHING OperationType = iota
	ROTATE
	SCALE
	TRANSLATE
)

const (
	sourceURL = "https://github.com/justinclift/tinygo_canvas_test1"
)

var (
	// The empty world space
	worldSpace []Object

	// The point objects
	axes = Object{
		C:         "grey",
		DrawOrder: 0,
		Name:      "axes",
		P: []Point{
			{X: -0.1, Y: 0.1, Z: 0.0},
			{X: -0.1, Y: 10, Z: 0.0},
			{X: 0.1, Y: 10, Z: 0.0},
			{X: 0.1, Y: 0.1, Z: 0.0},
			{X: 10, Y: 0.1, Z: 0.0},
			{X: 10, Y: -0.1, Z: 0.0},
			{X: 0.1, Y: -0.1, Z: 0.0},
			{X: 0.1, Y: -10, Z: 0.0},
			{X: -0.1, Y: -10, Z: 0.0},
			{X: -0.1, Y: -0.1, Z: 0.0},
			{X: -10, Y: -0.1, Z: 0.0},
			{X: -10, Y: 0.1, Z: 0.0},
			{X: 10, Y: -1.0, Z: 0.0, Label: "X", LabelAlign: "center"},
			{X: -10, Y: -1.0, Z: 0.0, Label: "-X", LabelAlign: "center"},
			{X: 0.0, Y: 10.5, Z: 0.0, Label: "Y", LabelAlign: "center"},
			{X: 0.0, Y: -11, Z: 0.0, Label: "-Y", LabelAlign: "center"},
		},
		E: []Edge{
			{0, 1},
			{1, 2},
			{2, 3},
			{3, 4},
			{4, 5},
			{5, 6},
			{6, 7},
			{7, 8},
			{8, 9},
			{9, 10},
			{10, 11},
			{11, 0},
		},
		S: []Surface{
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	// The 4x4 identity matrix
	identityMatrix = matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// Initialise the transform matrix with the identity matrix
	transformMatrix = identityMatrix

	canvasEl, ctx, doc js.Value
	graphWidth         float64
	graphHeight        float64
	height, width      int
	opText             string
	highLightSource    bool
	pointStep          = 0.05
	stepSize           = float64(25)

	// Queue operations
	prevKey    int
	queueOp    OperationType
	queueParts int32

	debug = false
)

func main() {
}

//go:export clearCanvas
func clearCanvas() {
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Add the X/Y axes object to the world space
	worldSpace = append(worldSpace, importObject(axes, 0.0, 0.0, 0.0))

	// Create a graph object with the main data points on it
	var firstDeriv, graph Object
	var p Point
	graphLabeled := false
	for x := -2.1; x <= 2.2; x += 0.05 {
		p = Point{X: x, Y: x * x * x} // y = x^3
		if !graphLabeled {
			p.Label = " Equation: y = x³ "
			p.LabelAlign = "right"
			graphLabeled = true
		}
		graph.P = append(graph.P, p)
	}
	graph.C = "blue"
	graph.DrawOrder = 1
	graph.Name = "graph"
	worldSpace = append(worldSpace, importObject(graph, 0.0, 0.0, 0.0))

	// Create a graph object with the 1st order derivative points on it
	graphLabeled = false
	for x := -2.1; x <= 2.2; x += pointStep {
		p = Point{X: x, Y: 3 * (x * x)} // y = 3x^2
		if !graphLabeled {
			p.Label = " 1st order derivative: y = 3x² "
			p.LabelAlign = "right"
			graphLabeled = true
		}
		firstDeriv.P = append(firstDeriv.P, p)
	}
	firstDeriv.C = "green"
	firstDeriv.DrawOrder = 2
	firstDeriv.Name = "firstDeriv"
	worldSpace = append(worldSpace, importObject(firstDeriv, 0.0, 0.0, 0.0))
}

// Simple mouse handler watching for people clicking on the source code link
//go:export clickHandler
func clickHandler(cx int, cy int) {
	clientX := float64(cx)
	clientY := float64(cy)
	if debug {
		println("ClientX: " + strconv.FormatFloat(clientX, 'f', 0, 64) + " clientY: " + strconv.FormatFloat(clientY, 'f', 0, 64))
		if clientX > graphWidth && clientY > (float64(height)-40) {
			println("URL hit!")
		}
	}

	// If the user clicks the source code URL area, open the URL
	if clientX > graphWidth && clientY > (float64(height)-40) {
		w := js.Global().Call("open", sourceURL)
		if w == js.Null() {
			// Couldn't open a new window, so try loading directly in the existing one instead
			doc.Set("location", sourceURL)
		}
	}
}

// Returns an object whose points have been transformed into 3D world space XYZ co-ordinates.  Also assigns a number
// to each point
//go:export importObject
func importObject(ob Object, x float64, y float64, z float64) (translatedObject Object) {
	// X and Y translation matrix.  Translates the objects into the world space at the given X and Y co-ordinates
	translateMatrix := matrix{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}

	// Translate the points
	var pt Point
	for _, j := range ob.P {
		pt = Point{
			Label:      j.Label,
			LabelAlign: j.LabelAlign,
			X:          (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:          (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:          (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		}
		translatedObject.P = append(translatedObject.P, pt)
	}

	// Copy the remaining object info across
	translatedObject.C = ob.C
	translatedObject.Name = ob.Name
	translatedObject.DrawOrder = ob.DrawOrder
	for _, j := range ob.E {
		translatedObject.E = append(translatedObject.E, j)
	}
	for _, j := range ob.S {
		translatedObject.S = append(translatedObject.S, j)
	}

	return translatedObject
}

// Simple keyboard handler for catching the arrow, WASD, and numpad keys
// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
//go:export keyPressHandler
func keyPressHandler(keyVal int) {
	if debug {
		println("Key is: " + strconv.Itoa(keyVal))
	}

	// If a key is pressed for a 2nd time in a row, then stop the animated movement
	if keyVal == prevKey && queueOp != NOTHING {
		queueOp = NOTHING
		return
	}

	// The the plus or minus keys were pressed, increase the step size then cause the current operation to be recalculated
	switch keyVal {
	case KEY_MINUS:
		stepSize -= 5.0
		keyVal = prevKey
	case KEY_PLUS:
		stepSize += 5.0
		keyVal = prevKey
	}

	// Set up translate and rotate operations
	switch keyVal {
	case KEY_MOVE_LEFT:
		setUpOperation(TRANSLATE, 50, 12, stepSize/2, 0, 0)
	case KEY_MOVE_RIGHT:
		setUpOperation(TRANSLATE, 50, 12, -stepSize/2, 0, 0)
	case KEY_MOVE_UP:
		setUpOperation(TRANSLATE, 50, 12, 0, stepSize/2, 0)
	case KEY_MOVE_DOWN:
		setUpOperation(TRANSLATE, 50, 12, 0, -stepSize/2, 0)
	case KEY_ROTATE_LEFT:
		setUpOperation(ROTATE, 50, 12, 0, -stepSize, 0)
	case KEY_ROTATE_RIGHT:
		setUpOperation(ROTATE, 50, 12, 0, stepSize, 0)
	case KEY_ROTATE_UP:
		setUpOperation(ROTATE, 50, 12, -stepSize, 0, 0)
	case KEY_ROTATE_DOWN:
		setUpOperation(ROTATE, 50, 12, stepSize, 0, 0)
	case KEY_PAGE_UP:
		setUpOperation(ROTATE, 50, 12, -stepSize, stepSize, 0)
	case KEY_PAGE_DOWN:
		setUpOperation(ROTATE, 50, 12, stepSize, stepSize, 0)
	case KEY_HOME:
		setUpOperation(ROTATE, 50, 12, -stepSize, -stepSize, 0)
	case KEY_END:
		setUpOperation(ROTATE, 50, 12, stepSize, -stepSize, 0)
	}
	prevKey = keyVal
}

// Set up the details for the transformation operation
//go:export setUpOperation
func setUpOperation(op OperationType, t int32, f int32, X float64, Y float64, Z float64) {
	queueParts = f                   // Number of parts to break each transformation into
	transformMatrix = identityMatrix // Reset the transform matrix
	switch op {
	case ROTATE: // Rotate the objects in world space
		// Divide the desired angle into a small number of parts
		if X != 0 {
			transformMatrix = rotateAroundX(transformMatrix, X/float64(queueParts))
		}
		if Y != 0 {
			transformMatrix = rotateAroundY(transformMatrix, Y/float64(queueParts))
		}
		if Z != 0 {
			transformMatrix = rotateAroundZ(transformMatrix, Z/float64(queueParts))
		}
		opText = "Rotation. X: " + strconv.FormatFloat(X, 'f', 0, 64) + " Y: " + strconv.FormatFloat(Y, 'f', 0, 64) + " Z: " + strconv.FormatFloat(Z, 'f', 0, 64)

	case SCALE:
		// Scale the objects in world space
		var xPart, yPart, zPart float64
		if X != 1 {
			xPart = ((X - 1) / float64(queueParts)) + 1
		}
		if Y != 1 {
			yPart = ((Y - 1) / float64(queueParts)) + 1
		}
		if Z != 1 {
			zPart = ((Z - 1) / float64(queueParts)) + 1
		}
		transformMatrix = scale(transformMatrix, xPart, yPart, zPart)
		opText = "Scale. X: " + strconv.FormatFloat(X, 'f', 0, 64) + " Y: " + strconv.FormatFloat(Y, 'f', 0, 64) + " Z: " + strconv.FormatFloat(Z, 'f', 0, 64)

	case TRANSLATE:
		// Translate (move) the objects in world space
		transformMatrix = translate(transformMatrix, X/float64(queueParts), Y/float64(queueParts), Z/float64(queueParts))
		opText = "Translate. X: " + strconv.FormatFloat(X, 'f', 0, 64) + " Y: " + strconv.FormatFloat(Y, 'f', 0, 64) + " Z: " + strconv.FormatFloat(Z, 'f', 0, 64)
	}
	queueOp = op
}

// Apply each transformation, one small part at a time (this gives the animation effect)
//go:export applyTransformation
func applyTransformation() {
	if (queueParts < 1 && queueOp == SCALE) || queueOp == NOTHING {
		opText = "Complete."
		return
	}

	// If the queue # if greater than zero, there are still transforms to do
	for j, o := range worldSpace {
		var newPoints []Point

		// Transform each point of in the object
		for _, j := range o.P {
			newPoints = append(newPoints, transform(transformMatrix, j))
		}
		o.P = newPoints

		// Update the object in world space
		worldSpace[j] = o
	}
	queueParts--
}

// Multiplies one matrix by another
//go:export matrixMult
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

// Simple mouse handler watching for people moving the mouse over the source code link
//go:export moveHandler
func moveHandler(cx int, cy int) {
	clientX := float64(cx)
	clientY := float64(cy)
	if debug {
		println("ClientX: " + strconv.FormatFloat(clientX, 'f', 0, 64) + " clientY: " + strconv.FormatFloat(clientY, 'f', 0, 64))
	}

	// If the mouse is over the source code link, let the frame renderer know to draw the url in bold
	if clientX > graphWidth && clientY > (float64(height)-40) {
		highLightSource = true
	} else {
		highLightSource = false
	}
}

// Renders one frame of the animation
//go:export renderFrame
func renderFrame() {
	// Handle window resizing
	curBodyW := js.Global().Get("innerWidth").Int()
	curBodyH := js.Global().Get("innerHeight").Int()
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
	centerX := graphWidth / 2
	centerY := graphHeight / 2

	// Clear the background
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", 0, 0, width, height)

	// Draw grid lines
	step := math.Min(float64(width), float64(height)) / float64(30)
	ctx.Set("strokeStyle", "rgb(220, 220, 220)")
	//ctx.Call("setLineDash", []interface{}{1, 3})
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

	// Draw the axes
	var pointX, pointY float64
	ctx.Set("strokeStyle", "black")
	ctx.Set("lineWidth", "1")
	//ctx.Call("setLineDash", []interface{}{})
	for _, o := range worldSpace {

		// Draw the surfaces
		ctx.Set("fillStyle", o.C)
		for _, l := range o.S {
			for m, n := range l {
				pointX = o.P[n].X
				pointY = o.P[n].Y
				if m == 0 {
					ctx.Call("beginPath")
					ctx.Call("moveTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
				} else {
					ctx.Call("lineTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
				}
			}
			ctx.Call("closePath")
			ctx.Call("fill")
		}

		// Draw the edges
		var point1X, point1Y, point2X, point2Y float64
		for _, l := range o.E {
			point1X = o.P[l[0]].X
			point1Y = o.P[l[0]].Y
			point2X = o.P[l[1]].X
			point2Y = o.P[l[1]].Y
			ctx.Call("beginPath")
			ctx.Call("moveTo", centerX+(point1X*step), centerY+((point1Y*step)*-1))
			ctx.Call("lineTo", centerX+(point2X*step), centerY+((point2Y*step)*-1))
			ctx.Call("stroke")
		}

		// Draw any point labels
		ctx.Set("fillStyle", "black")
		ctx.Set("font", "bold 16px serif")
		var px, py float64
		for _, l := range o.P {
			if l.Label != "" {
				ctx.Set("textAlign", l.LabelAlign)
				px = centerX + (l.X * step)
				py = centerY + ((l.Y * step) * -1)
				ctx.Call("fillText", l.Label, px, py)
			}
		}
	}

	// Draw the graph and derivatives
	ctx.Set("lineWidth", "2")
	//ctx.Call("setLineDash", []interface{}{})
	var px, py float64
	numWld := len(worldSpace)
	for i := 0; i < numWld; i++ {
		o := worldSpace[i]
		if o.Name != "axes" {
			// Draw lines between the points
			ctx.Set("strokeStyle", o.C)
			ctx.Call("beginPath")
			for k, l := range o.P {
				px = centerX + (l.X * step)
				py = centerY + ((l.Y * step) * -1)
				if k == 0 {
					ctx.Call("moveTo", px, py)
				} else {
					ctx.Call("lineTo", px, py)
				}
			}
			ctx.Call("stroke")

			// Draw dots for the points
			ctx.Set("fillStyle", "black")
			for _, l := range o.P {
				px = centerX + (l.X * step)
				py = centerY + ((l.Y * step) * -1)
				ctx.Call("beginPath")
				ctx.Call("ellipse", px, py, 1, 1, 0, 0, 2*math.Pi)
				ctx.Call("fill")
				ctx.Call("stroke")
			}
		}
	}

	// Clear the information area (right side)
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", graphWidth+1, 0, width, height)

	// Draw the text describing the current operation
	textY := top + 20
	ctx.Set("fillStyle", "black")
	ctx.Set("font", "bold 14px serif")
	ctx.Set("textAlign", "left")
	ctx.Call("fillText", "Operation:", graphWidth+20, textY)
	textY += 20
	ctx.Set("font", "14px sans-serif")
	ctx.Call("fillText", opText, graphWidth+20, textY)
	textY += 30

	// Add the help text about control keys and mouse zoom
	ctx.Set("fillStyle", "blue")
	ctx.Set("font", "14px sans-serif")
	ctx.Call("fillText", "Use wasd to move, numpad keys", graphWidth+20, textY)
	textY += 20
	ctx.Call("fillText", "to rotate, mouse wheel to zoom.", graphWidth+20, textY)
	textY += 30
	ctx.Call("fillText", "+ and - keys to change speed.", graphWidth+20, textY)
	textY += 30
	ctx.Call("fillText", "Press a key a 2nd time to", graphWidth+20, textY)
	textY += 20
	ctx.Call("fillText", "stop the current change.", graphWidth+20, textY)
	textY += 40

	// Add the graph and derivatives information
	ctx.Set("fillStyle", "black")
	ctx.Set("font", "bold 18px serif")
	ctx.Call("fillText", "Equation", graphWidth+20, textY)
	textY += 20
	ctx.Set("font", "16px sans-serif")
	ctx.Call("fillText", "y = x³", graphWidth+40, textY)
	textY += 30

	// Add the derivatives information
	ctx.Set("font", "bold 18px serif")
	ctx.Call("fillText", "1st order derivative", graphWidth+20, textY)
	textY += 20
	ctx.Set("font", "16px sans-serif")
	ctx.Call("fillText", "y = 3x²", graphWidth+40, textY)

	// Clear the source code link area
	ctx.Set("fillStyle", "white")
	ctx.Call("fillRect", graphWidth+1, graphHeight-55, width, height)

	// Add the URL to the source code
	ctx.Set("fillStyle", "black")
	ctx.Set("font", "bold 14px serif")
	ctx.Call("fillText", "Source code:", graphWidth+20, graphHeight-35)
	ctx.Set("fillStyle", "blue")
	if highLightSource == true {
		ctx.Set("font", "bold 12px sans-serif")
	} else {
		ctx.Set("font", "12px sans-serif")
	}
	ctx.Call("fillText", sourceURL, graphWidth+20, graphHeight-15)

	// Draw a border around the graph area
	//ctx.Call("setLineDash", []interface{}{})
	ctx.Set("lineWidth", "2")
	ctx.Set("strokeStyle", "white")
	ctx.Call("beginPath")
	ctx.Call("moveTo", 0, 0)
	ctx.Call("lineTo", width, 0)
	ctx.Call("lineTo", width, height)
	ctx.Call("lineTo", 0, height)
	ctx.Call("closePath")
	ctx.Call("stroke")
	ctx.Set("lineWidth", "2")
	ctx.Set("strokeStyle", "black")
	ctx.Call("beginPath")
	ctx.Call("moveTo", border, border)
	ctx.Call("lineTo", graphWidth, border)
	ctx.Call("lineTo", graphWidth, graphHeight)
	ctx.Call("lineTo", border, graphHeight)
	ctx.Call("closePath")
	ctx.Call("stroke")
}

// Rotates a transformation matrix around the X axis by the given degrees
//go:export rotateAroundX
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
//go:export rotateAroundY
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
//go:export rotateAroundZ
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
//go:export scale
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
//go:export transform
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
//go:export translate
func translate(m matrix, translateX float64, translateY float64, translateZ float64) matrix {
	translateMatrix := matrix{
		1, 0, 0, translateX,
		0, 1, 0, translateY,
		0, 0, 1, translateZ,
		0, 0, 0, 1,
	}
	return matrixMult(translateMatrix, m)
}

// Simple mouse handler watching for mouse wheel events
// Reference info can be found here: https://developer.mozilla.org/en-US/docs/Web/Events/wheel
//go:export wheelHandler
func wheelHandler(val int32) {
	wheelDelta := int64(val)
	scaleSize := 1 + (float64(wheelDelta) / 5)
	if debug {
		println("Wheel delta: " + strconv.FormatInt(wheelDelta, 10) + " scaleSize: " + strconv.FormatInt(wheelDelta, 10) + "\n")
	}
	setUpOperation(SCALE, 50, 12, scaleSize, scaleSize, scaleSize)
}
