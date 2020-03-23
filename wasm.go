package main

import (
	"math"
	"sort"
	"strconv"
	"syscall/js"
)

type matrix []float64

type Point struct {
	Num int
	X   float64
	Y   float64
	Z   float64
}

type Edge []int
type Surface []int

type Object struct {
	C   string // Colour of the object
	P   []Point
	E   []Edge    // List of points to connect by edges
	S   []Surface // List of points to connect in order, to create a surface
	Mid Point     // The mid point of the object.  Used for calculating object draw order in a very simple way
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

type paintOrder struct {
	midZ float64 // Z depth of an object's mid point
	name string
}

type paintOrderSlice []paintOrder

func (p paintOrder) String() string {
	return "Name: " + p.name + ", Mid point: " + strconv.FormatFloat(p.midZ, 'f', 1, 64)
}

func (p paintOrderSlice) Len() int {
	return len(p)
}

func (p paintOrderSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p paintOrderSlice) Less(i, j int) bool {
	return p[i].midZ < p[j].midZ
}

const sourceURL = "https://github.com/justinclift/tinygo_canvas2"

var (
	// The empty world space
	worldSpace   map[string]Object
	pointCounter = 1

	// The point objects
	object1 = Object{
		C: "lightblue",
		P: []Point{
			{X: 0, Y: 1.75, Z: 1.0},    // Point 0 for this object
			{X: 1.5, Y: -1.75, Z: 1.0}, // Point 1 for this object
			{X: -1.5, Y: -1.75, Z: 1.0},
			{X: 0, Y: 0, Z: 1.75},
		},
		E: []Edge{
			{0, 1}, // Connect point 0 to point 1
			{0, 2}, // Connect point 0 to point 2
			{1, 2}, // Connect point 1 to point 2
			{0, 3}, // etc
			{1, 3},
			{2, 3},
		},
		S: []Surface{
			{0, 1, 3},
			{0, 2, 3},
			{0, 1, 2},
			{1, 2, 3},
		},
	}
	object2 = Object{
		C: "lightgreen",
		P: []Point{
			{X: 1.5, Y: 1.5, Z: -1.0},  // Point 0 for this object
			{X: 1.5, Y: -1.5, Z: -1.0}, // Point 1 for this object
			{X: -1.5, Y: -1.5, Z: -1.0},
		},
		E: []Edge{
			{0, 1}, // Connect point 0 to point 1
			{1, 2}, // Connect point 1 to point 2
			{2, 0}, // etc
		},
		S: []Surface{
			{0, 1, 2},
		},
	}
	object3 = Object{
		C: "indianred",
		P: []Point{
			{X: 2, Y: -2, Z: 1.0},
			{X: 2, Y: -4, Z: 1.0},
			{X: -2, Y: -4, Z: 1.0},
			{X: -2, Y: -2, Z: 1.0},
			{X: 0, Y: -3, Z: 2.5},
		},
		E: []Edge{
			{0, 1},
			{1, 2},
			{2, 3},
			{3, 0},
			{0, 4},
			{1, 4},
			{2, 4},
			{3, 4},
		},
		S: []Surface{
			{0, 1, 4},
			{1, 2, 4},
			{2, 3, 4},
			{3, 0, 4},
			{0, 1, 2, 3},
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
	width, height      float64
	opText             string
	highLightSource    bool
	stepSize           = float64(15)

	// Queue operations
	prevKey    int
	queueOp    OperationType
	queueParts int32

	debug = true
)

func main() {
	width := js.Global().Get("innerWidth").Int()
	height := js.Global().Get("innerHeight").Int()
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Add some objects to the world space
	worldSpace = make(map[string]Object, 1)
	worldSpace["ob1"] = importObject(object1, 5.0, 3.0, 0.0)
	worldSpace["ob1 copy"] = importObject(object1, -1.0, 3.0, 0.0)
	worldSpace["ob2"] = importObject(object2, 5.0, -3.0, 1.0)
	worldSpace["ob3"] = importObject(object3, -1.0, 0.0, -1.0)

	// Scale them up a bit
	queueOp = SCALE
	queueParts = 1
	transformMatrix = scale(transformMatrix, 2.0, 2.0, 2.0)
	applyTransformation()

	// Start a rotation going
	setUpOperation(ROTATE, 50, 12, stepSize, stepSize, stepSize)

	// Start the frame renderer
	js.Global().Call("requestAnimationFrame", js.Global().Get("renderFrame"))
}

// Apply each transformation, one small part at a time (this gives the animation effect)
//go:export applyTransformation
func applyTransformation() {
	if (queueParts < 1 && queueOp == SCALE) || queueOp == NOTHING {
		opText = "Complete."
		return
	}

	for j, o := range worldSpace {
		var newPoints []Point
		// Transform each point in the object
		for _, j := range o.P {
			newPoints = append(newPoints, transform(transformMatrix, j))
		}
		o.P = newPoints

		// Transform the mid point of the object.  In theory, this should mean the mid point can always be used
		// for a simple (not-cpu-intensive) way to sort the objects in Z depth order
		o.Mid = transform(transformMatrix, o.Mid)

		// Update the object in world space
		worldSpace[j] = o
	}

	queueParts--
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
	curBodyW := js.Global().Get("innerWidth").Float()
	curBodyH := js.Global().Get("innerHeight").Float()
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

	// Save the current graphics state - no clip region currently defined - as the default
	ctx.Call("save")

	// Set the clip region so drawing only occurs in the display area
	ctx.Call("beginPath")
	ctx.Call("moveTo", 0, 0)
	ctx.Call("lineTo", graphWidth, 0)
	ctx.Call("lineTo", graphWidth, height)
	ctx.Call("lineTo", 0, height)
	ctx.Call("clip")

	// Draw grid lines
	step := math.Min(float64(width), float64(height)) / float64(30)
	ctx.Set("strokeStyle", "rgb(220, 220, 220)")
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

	// Sort the objects by mid point Z depth order
	var order paintOrderSlice
	for i, j := range worldSpace {
		order = append(order, paintOrder{name: i, midZ: j.Mid.Z})
	}
	sort.Sort(paintOrderSlice(order))

	// Draw the objects, in Z depth order
	var pointX, pointY float64
	numWld := len(worldSpace)
	for i := 0; i < numWld; i++ {
		o := worldSpace[order[i].name]

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
		ctx.Set("strokeStyle", "black")
		ctx.Set("fillStyle", "black")
		ctx.Set("lineWidth", "1")
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

		// Draw the points on the graph
		var px, py float64
		for _, l := range o.P {
			px = centerX + (l.X * step)
			py = centerY + ((l.Y * step) * -1)
			ctx.Call("beginPath")
			ctx.Call("arc", px, py, 1, 0, 2*math.Pi)
			ctx.Call("fill")
		}
	}

	// Set the clip region so drawing only occurs in the display area
	ctx.Call("restore")
	ctx.Call("save")
	ctx.Call("beginPath")
	ctx.Call("moveTo", graphWidth, 0)
	ctx.Call("lineTo", width, 0)
	ctx.Call("lineTo", width, height)
	ctx.Call("lineTo", graphWidth, height)
	ctx.Call("clip")

	// Draw the text describing the current operation
	textY := top + 20
	ctx.Set("fillStyle", "black")
	ctx.Set("font", "bold 14px serif")
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

	// Restore the default graphics state (eg no clip region)
	ctx.Call("restore")

	// Keep the frame rendering going
	js.Global().Call("requestAnimationFrame", js.Global().Get("renderFrame"))
}

// Simple mouse handler watching for mouse wheel events
// Reference info can be found here: https://developer.mozilla.org/en-US/docs/Web/Events/wheel
//go:export wheelHandler
func wheelHandler(val int32) {
	wheelDelta := int64(val)
	scaleSize := 1 + (float64(wheelDelta) / 5)
	if debug {
		println("Wheel delta: " + strconv.FormatInt(wheelDelta, 10) + " scaleSize: " + strconv.FormatFloat(scaleSize, 'f', 1, 64) + "\n")
	}
	setUpOperation(SCALE, 50, 12, scaleSize, scaleSize, scaleSize)
}

// Returns an object whose points have been transformed into 3D world space XYZ co-ordinates.  Also assigns a number
// to each point
func importObject(ob Object, x float64, y float64, z float64) (translatedObject Object) {
	// X and Y translation matrix.  Translates the objects into the world space at the given X and Y co-ordinates
	translateMatrix := matrix{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}

	// Translate the points
	var midX, midY, midZ float64
	var pt Point
	for _, j := range ob.P {
		pt = Point{
			Num: pointCounter,
			X:   (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:   (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:   (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		}
		translatedObject.P = append(translatedObject.P, pt)
		midX += pt.X
		midY += pt.Y
		midZ += pt.Z
		pointCounter++
	}

	// Determine the mid point for the object
	numPts := float64(len(ob.P))
	translatedObject.Mid.X = midX / numPts
	translatedObject.Mid.Y = midY / numPts
	translatedObject.Mid.Z = midZ / numPts

	// Copy the colour, edge, and surface definitions across
	translatedObject.C = ob.C
	for _, j := range ob.E {
		translatedObject.E = append(translatedObject.E, j)
	}
	for _, j := range ob.S {
		translatedObject.S = append(translatedObject.S, j)
	}

	return translatedObject
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

// Set up the details for the transformation operation
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

	t.Num = p.Num
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
