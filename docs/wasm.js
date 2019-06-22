'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function clearCanvas() {
  wasm.exports.clearCanvas();
}

// Pass mouse clicks through to the wasm handler
function clickHandler(evt) {
  wasm.exports.clickHandler(evt.clientX, evt.clientY);
}

// Pass key presses through to the wasm handler
function keyPressHandler(evt) {
  let key = 0;
  switch(evt.key) {
    // Move keys
    case "d":
    case "D":
      key = 1;
      break;
    case "a":
    case "A":
      key = 2;
      break;
    case "w":
    case "W":
      key = 3;
      break;
    case "s":
    case "S":
      key = 4;
      break;

    // Rotate keys
    case "ArrowLeft":
    case "4":
      key = 5;
      break;
    case "ArrowRight":
    case "6":
      key = 6;
      break;
    case "ArrowUp":
    case "8":
      key = 7;
      break;
    case "ArrowDown":
    case "2":
      key = 8;
      break;
    case "PageUp":
    case "9":
      key = 9;
      break;
    case "PageDown":
    case "3":
      key = 10;
      break;
    case "Home":
    case "7":
      key = 11;
      break;
    case "End":
    case "1":
      key = 12;
      break;

    // Change step size keys
    case "-":
      key = 13;
      break;
    case "+":
      key = 14;
      break;

    // Unknown key press, don't pass it through
    default:
      return;
  }

  // console.log("JS: Key pressed = " + key);
  wasm.exports.keyPressHandler(key);
}

// Pass mouse movement events through to its wasm handler
function moveHandler(evt) {
  // console.log(evt);
  wasm.exports.moveHandler(evt.clientX, evt.clientY);
}

function renderFrames() {
    wasm.exports.applyTransformation();
    wasm.exports.renderFrame();
}

// Pass mouse wheel events through to its wasm handler
function wheelHandler(evt) {
    wasm.exports.wheelHandler(evt.deltaY);
}


function init() {
  const go = new Go();
  if ('instantiateStreaming' in WebAssembly) {
    WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
      wasm = obj.instance;
      go.run(wasm);

      // Set up wasm event handlers
      document.getElementById("mycanvas").addEventListener("mousedown", clickHandler);
      document.getElementById("mycanvas").addEventListener("keydown", keyPressHandler);
      document.getElementById("mycanvas").addEventListener("mousemove", moveHandler);
      document.getElementById("mycanvas").addEventListener("wheel", wheelHandler);

      // Set up the canvas
      clearCanvas();

      // Set up basic render loop
      setInterval(function() {
        renderFrames();
      },50);
    })
  } else {
    fetch(WASM_URL).then(resp =>
      resp.arrayBuffer()
    ).then(bytes =>
      WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);

        // Set up wasm event handlers
        document.getElementById("mycanvas").addEventListener("mousedown", clickHandler);
        document.getElementById("mycanvas").addEventListener("keydown", keyPressHandler);
        document.getElementById("mycanvas").addEventListener("mousemove", moveHandler);
        document.getElementById("mycanvas").addEventListener("wheel", wheelHandler);

        // Set up the canvas
        clearCanvas();

        // Set up basic render loop
        setInterval(function() {
          renderFrames();
        },50);
      })
    )
  }
}

init();
