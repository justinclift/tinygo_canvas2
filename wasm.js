'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function clearCanvas() {
  wasm.exports.clearCanvas();
}

function mouseDownHandler(evtDetails) {
  wasm.exports.mouseDownHandler(evtDetails.clientX, evtDetails.clientY);
}

function drawLine() {
  wasm.exports.drawLine();
}

function init() {
  const go = new Go();
  if ('instantiateStreaming' in WebAssembly) {
    WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
      wasm = obj.instance;
      go.run(wasm);

      // Set up mouse click handler
      document.getElementById("mycanvas").addEventListener("mousedown", mouseDownHandler);

      clearCanvas();
    })
  } else {
    fetch(WASM_URL).then(resp =>
      resp.arrayBuffer()
    ).then(bytes =>
      WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);

        // Set up mouse click handler
        document.getElementById("mycanvas").addEventListener("mousedown", mouseDownHandler);

        clearCanvas();
      })
    )
  }
}

init();
