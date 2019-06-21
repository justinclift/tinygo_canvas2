'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function clearCanvas() {
  wasm.exports.clearCanvas();
}

// Pass mouse clicks through to the wasm handler
function clickHandler(evt) {
  // console.log(evt);
  wasm.exports.clickHandler(evt.clientX, evt.clientY);
}

// Pass key presses through to the wasm handler
function keyPressHandler(evt) {
  // console.log(evt);
  let key = 0;
  switch(evt.key) {
    case "ArrowLeft":
    case "a":
    case "A":
    case "4":
      key = 1;
      break;
    case "ArrowRight":
    case "d":
    case "D":
    case "6":
      key = 2;
      break;
    case "ArrowUp":
    case "w":
    case "W":
    case "8":
      key = 3;
      break;
    case "ArrowDown":
    case "s":
    case "S":
    case "2":
      key = 4;
      break;
    case "PageUp":
    case "9":
      key = 5;
      break;
    case "PageDown":
    case "3":
      key = 6;
      break;
    case "Home":
    case "7":
      key = 7;
      break;
    case "End":
    case "1":
      key = 8;
      break;
    case "-":
      key = 9;
      break;
    case "+":
      key = 10;
      break;
    default:
      // Unknown key press, don't pass it through
      return;
  }

  // console.log("JS: Key pressed = " + key);
  wasm.exports.keyPressHandler(key);
}

// Pass mouse movement events through to the wasm handler
function moveHandler(evt) {
  // console.log(evt);
  wasm.exports.moveHandler(evt.clientX, evt.clientY);
}

var framesRendered = 0;
function renderFrames() {
  // if (framesRendered <= 1) {
    console.log("Frames rendered: " + framesRendered);
    framesRendered = framesRendered + 1;
    wasm.exports.renderFrame();
  // }
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
